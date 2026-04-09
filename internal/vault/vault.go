package vault

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nitzzzu/hush/internal/crypto"
	"github.com/nitzzzu/hush/internal/keychain"
	_ "modernc.org/sqlite"
)

const maxHistory = 10

// Secret represents a stored secret.
type Secret struct {
	ID        int64
	Name      string
	Value     string
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// HistoryEntry represents a historical version of a secret.
type HistoryEntry struct {
	ID         int64
	SecretName string
	Version    int
	CreatedAt  time.Time
}

// Vault manages the SQLite-backed secrets store.
type Vault struct {
	db        *sql.DB
	key       []byte
	vaultPath string
}

// Open opens an existing vault at the given path.
func Open(vaultPath string) (*Vault, error) {
	key, err := keychain.Get(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}
	db, err := sql.Open("sqlite", vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open vault: %w", err)
	}
	v := &Vault{db: db, key: key, vaultPath: vaultPath}
	if err := v.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return v, nil
}

// Init creates a new vault at the given path, generates a key, and stores it in the keychain.
func Init(vaultPath string) (*Vault, error) {
	dir := filepath.Dir(vaultPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create vault directory: %w", err)
	}
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	if err := keychain.Set(vaultPath, key); err != nil {
		return nil, fmt.Errorf("failed to store key: %w", err)
	}
	db, err := sql.Open("sqlite", vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault: %w", err)
	}
	v := &Vault{db: db, key: key, vaultPath: vaultPath}
	if err := v.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return v, nil
}

// Close closes the vault database connection.
func (v *Vault) Close() error {
	return v.db.Close()
}

// Path returns the vault's file path.
func (v *Vault) Path() string {
	return v.vaultPath
}

func (v *Vault) migrate() error {
	_, err := v.db.Exec(`
		CREATE TABLE IF NOT EXISTS secrets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			encrypted_value BLOB NOT NULL,
			tags TEXT NOT NULL DEFAULT '[]',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
		CREATE TABLE IF NOT EXISTS secret_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			secret_name TEXT NOT NULL,
			encrypted_value BLOB NOT NULL,
			version INTEGER NOT NULL,
			created_at DATETIME NOT NULL
		);
	`)
	return err
}

// SetSecret encrypts and stores a secret value.
func (v *Vault) SetSecret(name, value string, tags []string) error {
	if tags == nil {
		tags = []string{}
	}
	encrypted, err := crypto.Encrypt(v.key, []byte(value))
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// Check if exists — if so, archive current value
	existing, err := v.getEncryptedValue(name)
	if err == nil {
		// Archive current before overwrite
		if archiveErr := v.archiveVersion(name, existing); archiveErr != nil {
			return archiveErr
		}
		// Update
		_, err = v.db.Exec(
			`UPDATE secrets SET encrypted_value=?, tags=?, updated_at=? WHERE name=?`,
			encrypted, string(tagsJSON), now, name,
		)
		return err
	}

	// Insert new
	_, err = v.db.Exec(
		`INSERT INTO secrets (name, encrypted_value, tags, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		name, encrypted, string(tagsJSON), now, now,
	)
	return err
}

func (v *Vault) getEncryptedValue(name string) ([]byte, error) {
	var enc []byte
	err := v.db.QueryRow(`SELECT encrypted_value FROM secrets WHERE name=?`, name).Scan(&enc)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func (v *Vault) archiveVersion(name string, encryptedValue []byte) error {
	var maxVer sql.NullInt64
	v.db.QueryRow(`SELECT MAX(version) FROM secret_history WHERE secret_name=?`, name).Scan(&maxVer)
	nextVer := int64(1)
	if maxVer.Valid {
		nextVer = maxVer.Int64 + 1
	}

	_, err := v.db.Exec(
		`INSERT INTO secret_history (secret_name, encrypted_value, version, created_at) VALUES (?, ?, ?, ?)`,
		name, encryptedValue, nextVer, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	// Prune to keep only last maxHistory versions
	_, err = v.db.Exec(`
		DELETE FROM secret_history
		WHERE secret_name=? AND id NOT IN (
			SELECT id FROM secret_history WHERE secret_name=? ORDER BY version DESC LIMIT ?
		)
	`, name, name, maxHistory)
	return err
}

// GetSecret retrieves and decrypts a secret by name.
func (v *Vault) GetSecret(name string) (*Secret, error) {
	var s Secret
	var encValue []byte
	var tagsJSON string
	var createdAt, updatedAt string

	err := v.db.QueryRow(
		`SELECT id, name, encrypted_value, tags, created_at, updated_at FROM secrets WHERE name=?`,
		name,
	).Scan(&s.ID, &s.Name, &encValue, &tagsJSON, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("secret %q not found", name)
	}
	if err != nil {
		return nil, err
	}

	decrypted, err := crypto.Decrypt(v.key, encValue)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret: %w", err)
	}
	s.Value = string(decrypted)

	if err := json.Unmarshal([]byte(tagsJSON), &s.Tags); err != nil {
		s.Tags = []string{}
	}

	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &s, nil
}

// ListSecrets returns all secrets, optionally filtered by tags.
func (v *Vault) ListSecrets(filterTags []string) ([]*Secret, error) {
	rows, err := v.db.Query(
		`SELECT id, name, encrypted_value, tags, created_at, updated_at FROM secrets ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []*Secret
	for rows.Next() {
		var s Secret
		var encValue []byte
		var tagsJSON string
		var createdAt, updatedAt string

		if err := rows.Scan(&s.ID, &s.Name, &encValue, &tagsJSON, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(tagsJSON), &s.Tags); err != nil {
			s.Tags = []string{}
		}
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		if len(filterTags) > 0 && !hasTags(s.Tags, filterTags) {
			continue
		}

		decrypted, err := crypto.Decrypt(v.key, encValue)
		if err != nil {
			return nil, err
		}
		s.Value = string(decrypted)
		secrets = append(secrets, &s)
	}
	return secrets, rows.Err()
}

// AllSecrets returns all secrets with decrypted values (for injection).
func (v *Vault) AllSecrets() ([]*Secret, error) {
	return v.ListSecrets(nil)
}

// DeleteSecret removes a secret by name.
func (v *Vault) DeleteSecret(name string) error {
	result, err := v.db.Exec(`DELETE FROM secrets WHERE name=?`, name)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("secret %q not found", name)
	}
	// Also delete history
	_, err = v.db.Exec(`DELETE FROM secret_history WHERE secret_name=?`, name)
	return err
}

// GetHistory returns the version history for a secret.
func (v *Vault) GetHistory(name string) ([]HistoryEntry, error) {
	rows, err := v.db.Query(
		`SELECT id, secret_name, version, created_at FROM secret_history WHERE secret_name=? ORDER BY version DESC`,
		name,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []HistoryEntry
	for rows.Next() {
		var h HistoryEntry
		var createdAt string
		if err := rows.Scan(&h.ID, &h.SecretName, &h.Version, &createdAt); err != nil {
			return nil, err
		}
		h.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		history = append(history, h)
	}
	return history, rows.Err()
}

// Rollback restores a secret to the specified version.
func (v *Vault) Rollback(name string, version int) error {
	var encValue []byte
	err := v.db.QueryRow(
		`SELECT encrypted_value FROM secret_history WHERE secret_name=? AND version=?`,
		name, version,
	).Scan(&encValue)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("version %d not found for secret %q", version, name)
	}
	if err != nil {
		return err
	}

	// Archive current value before rollback
	current, err := v.getEncryptedValue(name)
	if err == nil {
		_ = v.archiveVersion(name, current)
	}

	_, err = v.db.Exec(
		`UPDATE secrets SET encrypted_value=?, updated_at=? WHERE name=?`,
		encValue, time.Now().UTC().Format(time.RFC3339), name,
	)
	return err
}

// AddTag adds a tag to a secret.
func (v *Vault) AddTag(name, tag string) error {
	var tagsJSON string
	err := v.db.QueryRow(`SELECT tags FROM secrets WHERE name=?`, name).Scan(&tagsJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("secret %q not found", name)
	}
	if err != nil {
		return err
	}

	var tags []string
	_ = json.Unmarshal([]byte(tagsJSON), &tags)
	for _, t := range tags {
		if t == tag {
			return nil
		}
	}
	tags = append(tags, tag)

	updated, _ := json.Marshal(tags)
	_, err = v.db.Exec(`UPDATE secrets SET tags=?, updated_at=? WHERE name=?`, string(updated), time.Now().UTC().Format(time.RFC3339), name)
	return err
}

// RemoveTag removes a tag from a secret.
func (v *Vault) RemoveTag(name, tag string) error {
	var tagsJSON string
	err := v.db.QueryRow(`SELECT tags FROM secrets WHERE name=?`, name).Scan(&tagsJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("secret %q not found", name)
	}
	if err != nil {
		return err
	}

	var tags []string
	_ = json.Unmarshal([]byte(tagsJSON), &tags)
	newTags := make([]string, 0, len(tags))
	for _, t := range tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}

	updated, _ := json.Marshal(newTags)
	_, err = v.db.Exec(`UPDATE secrets SET tags=?, updated_at=? WHERE name=?`, string(updated), time.Now().UTC().Format(time.RFC3339), name)
	return err
}

func hasTags(secretTags, filterTags []string) bool {
	tagSet := make(map[string]bool)
	for _, t := range secretTags {
		tagSet[t] = true
	}
	for _, t := range filterTags {
		if tagSet[t] {
			return true
		}
	}
	return false
}

// VaultPath returns the resolved vault path for the given options.
func VaultPath(global bool, envName string) (string, error) {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if envName != "" {
			return filepath.Join(home, ".hush", "envs", envName, "vault.db"), nil
		}
		return filepath.Join(home, ".hush", "vault.db"), nil
	}
	if envName != "" {
		return filepath.Join(".hush", "envs", envName, "vault.db"), nil
	}
	return filepath.Join(".hush", "vault.db"), nil
}

// ListEnvs returns the list of available environments in the vault directory.
func ListEnvs(global bool) ([]string, error) {
	var envsDir string
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		envsDir = filepath.Join(home, ".hush", "envs")
	} else {
		envsDir = filepath.Join(".hush", "envs")
	}

	entries, err := os.ReadDir(envsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var envs []string
	for _, e := range entries {
		if e.IsDir() {
			envs = append(envs, e.Name())
		}
	}
	return envs, nil
}
