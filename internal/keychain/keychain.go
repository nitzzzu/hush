package keychain

import (
	"crypto/sha256"
	"os"

	"github.com/99designs/keyring"
)

const serviceName = "hush"

func openKeyring() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName: serviceName,
		AllowedBackends: []keyring.BackendType{
			keyring.SecretServiceBackend,
			keyring.KWalletBackend,
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.FileBackend,
		},
		FileDir:          os.ExpandEnv("$HOME/.hush/keyring"),
		FilePasswordFunc: keyring.TerminalPrompt,
	})
}

// deriveKeyFromPassword derives a 32-byte key from password+salt using SHA-256.
func deriveKeyFromPassword(password, salt string) []byte {
	h := sha256.New()
	h.Write([]byte(password))
	h.Write([]byte(":"))
	h.Write([]byte(salt))
	return h.Sum(nil)
}

// Set stores a key in the OS keychain under vaultPath.
func Set(vaultPath string, key []byte) error {
	if os.Getenv("HUSH_PASSWORD") != "" {
		return nil
	}
	ring, err := openKeyring()
	if err != nil {
		return err
	}
	return ring.Set(keyring.Item{
		Key:  vaultPath,
		Data: key,
	})
}

// Get retrieves a key from the OS keychain.
// Falls back to HUSH_PASSWORD env var.
func Get(vaultPath string) ([]byte, error) {
	if pw := os.Getenv("HUSH_PASSWORD"); pw != "" {
		return deriveKeyFromPassword(pw, vaultPath), nil
	}
	ring, err := openKeyring()
	if err != nil {
		return nil, err
	}
	item, err := ring.Get(vaultPath)
	if err != nil {
		return nil, err
	}
	return item.Data, nil
}

// Delete removes a key from the OS keychain.
func Delete(vaultPath string) error {
	if os.Getenv("HUSH_PASSWORD") != "" {
		return nil
	}
	ring, err := openKeyring()
	if err != nil {
		return err
	}
	return ring.Remove(vaultPath)
}
