package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	assert.NoError(t, err)
	assert.Len(t, key, 32)

	key2, err := GenerateKey()
	assert.NoError(t, err)
	assert.NotEqual(t, key, key2)
}

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	assert.NoError(t, err)

	plaintext := []byte("super secret value")
	ciphertext, err := Encrypt(key, plaintext)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := Decrypt(key, ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestDecryptWrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	ciphertext, err := Encrypt(key1, []byte("secret"))
	assert.NoError(t, err)

	_, err = Decrypt(key2, ciphertext)
	assert.Error(t, err)
}

func TestDecryptTooShort(t *testing.T) {
	key, _ := GenerateKey()
	_, err := Decrypt(key, []byte("short"))
	assert.Error(t, err)
}

func TestEncryptProducesUniqueNonces(t *testing.T) {
	key, _ := GenerateKey()
	plaintext := []byte("same input")

	c1, _ := Encrypt(key, plaintext)
	c2, _ := Encrypt(key, plaintext)
	assert.NotEqual(t, c1, c2)
}
