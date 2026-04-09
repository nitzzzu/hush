package vault

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestVault(t *testing.T) (*Vault, string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HUSH_PASSWORD", "test-password")
	vaultPath := filepath.Join(dir, "vault.db")
	v, err := Init(vaultPath)
	require.NoError(t, err)
	t.Cleanup(func() { v.Close() })
	return v, vaultPath
}

func TestInitAndOpen(t *testing.T) {
	t.Setenv("HUSH_PASSWORD", "test-password")
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.db")

	v, err := Init(vaultPath)
	require.NoError(t, err)
	v.Close()

	v2, err := Open(vaultPath)
	require.NoError(t, err)
	defer v2.Close()
}

func TestSetAndGetSecret(t *testing.T) {
	v, _ := setupTestVault(t)

	err := v.SetSecret("MY_KEY", "my_value", []string{"production"})
	require.NoError(t, err)

	s, err := v.GetSecret("MY_KEY")
	require.NoError(t, err)
	assert.Equal(t, "MY_KEY", s.Name)
	assert.Equal(t, "my_value", s.Value)
	assert.Equal(t, []string{"production"}, s.Tags)
}

func TestGetSecretNotFound(t *testing.T) {
	v, _ := setupTestVault(t)
	_, err := v.GetSecret("NONEXISTENT")
	assert.Error(t, err)
}

func TestUpdateSecret(t *testing.T) {
	v, _ := setupTestVault(t)

	err := v.SetSecret("KEY", "value1", nil)
	require.NoError(t, err)

	err = v.SetSecret("KEY", "value2", nil)
	require.NoError(t, err)

	s, err := v.GetSecret("KEY")
	require.NoError(t, err)
	assert.Equal(t, "value2", s.Value)
}

func TestHistoryTracking(t *testing.T) {
	v, _ := setupTestVault(t)

	require.NoError(t, v.SetSecret("KEY", "v1", nil))
	require.NoError(t, v.SetSecret("KEY", "v2", nil))
	require.NoError(t, v.SetSecret("KEY", "v3", nil))

	history, err := v.GetHistory("KEY")
	require.NoError(t, err)
	assert.Len(t, history, 2) // v1 and v2 are archived; v3 is current
}

func TestHistoryMaxVersions(t *testing.T) {
	v, _ := setupTestVault(t)

	for i := 0; i < 12; i++ {
		require.NoError(t, v.SetSecret("KEY", fmt.Sprintf("value%d", i), nil))
	}

	history, err := v.GetHistory("KEY")
	require.NoError(t, err)
	assert.LessOrEqual(t, len(history), maxHistory)
}

func TestDeleteSecret(t *testing.T) {
	v, _ := setupTestVault(t)

	require.NoError(t, v.SetSecret("KEY", "value", nil))
	err := v.DeleteSecret("KEY")
	require.NoError(t, err)

	_, err = v.GetSecret("KEY")
	assert.Error(t, err)
}

func TestListSecrets(t *testing.T) {
	v, _ := setupTestVault(t)

	require.NoError(t, v.SetSecret("A", "a", []string{"tag1"}))
	require.NoError(t, v.SetSecret("B", "b", []string{"tag2"}))
	require.NoError(t, v.SetSecret("C", "c", []string{"tag1", "tag2"}))

	all, err := v.ListSecrets(nil)
	require.NoError(t, err)
	assert.Len(t, all, 3)

	filtered, err := v.ListSecrets([]string{"tag1"})
	require.NoError(t, err)
	assert.Len(t, filtered, 2)
}

func TestAddRemoveTag(t *testing.T) {
	v, _ := setupTestVault(t)

	require.NoError(t, v.SetSecret("KEY", "value", nil))
	err := v.AddTag("KEY", "prod")
	require.NoError(t, err)

	s, err := v.GetSecret("KEY")
	require.NoError(t, err)
	assert.Contains(t, s.Tags, "prod")

	err = v.RemoveTag("KEY", "prod")
	require.NoError(t, err)

	s, err = v.GetSecret("KEY")
	require.NoError(t, err)
	assert.NotContains(t, s.Tags, "prod")
}

func TestRollback(t *testing.T) {
	v, _ := setupTestVault(t)

	require.NoError(t, v.SetSecret("KEY", "original", nil))
	require.NoError(t, v.SetSecret("KEY", "updated", nil))

	history, err := v.GetHistory("KEY")
	require.NoError(t, err)
	require.NotEmpty(t, history)

	err = v.Rollback("KEY", history[0].Version)
	require.NoError(t, err)

	s, err := v.GetSecret("KEY")
	require.NoError(t, err)
	assert.Equal(t, "original", s.Value)
}
