package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

const encPrefix = "enc:"

// IsEncrypted reports whether s was produced by Encrypt.
func IsEncrypted(s string) bool {
	return strings.HasPrefix(s, encPrefix)
}

// Encrypt encrypts plaintext with AES-256-GCM and returns "enc:<base64(nonce+ciphertext)>".
func Encrypt(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return encPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt decrypts a value produced by Encrypt. If s does not have the "enc:" prefix it is
// returned unchanged, so plain-text values in the config continue to work.
func Decrypt(key []byte, s string) (string, error) {
	if !IsEncrypted(s) {
		return s, nil
	}
	if key == nil {
		return "", fmt.Errorf("value is encrypted but ENCRYPTION_KEY is not set")
	}
	data, err := base64.StdEncoding.DecodeString(s[len(encPrefix):])
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: wrong key or corrupted data")
	}
	return string(plaintext), nil
}

// LoadKey reads the encryption key from the ENCRYPTION_KEY environment variable.
// The value must be exactly 32 bytes encoded as 64 hex chars or 44 base64 chars.
// Returns (nil, nil) when the variable is not set.
func LoadKey() ([]byte, error) {
	v := os.Getenv("ENCRYPTION_KEY")
	if v == "" {
		return nil, nil
	}
	if b, err := hex.DecodeString(v); err == nil {
		if len(b) != 32 {
			return nil, fmt.Errorf("ENCRYPTION_KEY hex value must be 32 bytes (64 hex chars), got %d", len(b))
		}
		return b, nil
	}
	b, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be 32-byte value encoded as 64 hex chars or base64")
	}
	if len(b) != 32 {
		return nil, fmt.Errorf("ENCRYPTION_KEY base64 value must decode to 32 bytes, got %d", len(b))
	}
	return b, nil
}
