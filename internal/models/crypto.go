package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "backlog-manager"
	keyringKey     = "encryption-key"
)

// getOrCreateKey retrieves or generates a persistent AES-256 key.
// It first tries to load it from the system keychain; if that fails, it falls back to a local file.
func getOrCreateKey() ([]byte, error) {
	// Try system keychain first
	keyStr, err := keyring.Get(keyringService, keyringKey)
	if err == nil && keyStr != "" {
		decoded, err := base64.StdEncoding.DecodeString(keyStr)
		if err == nil && len(decoded) == 32 {
			return decoded, nil
		}
	}

	// Fallback: local encrypted key file
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	keyDir := filepath.Join(configDir, "backlog-manager")
	os.MkdirAll(keyDir, 0700)
	keyPath := filepath.Join(keyDir, "key.bin")

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}

		// Try saving to keychain
		_ = keyring.Set(keyringService, keyringKey, base64.StdEncoding.EncodeToString(key))

		// Always also persist locally as fallback
		if err := os.WriteFile(keyPath, key, 0600); err != nil {
			return nil, err
		}
		return key, nil
	}

	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// Encrypt secures a plaintext string using AES-256-GCM and returns a base64-encoded ciphertext.
func Encrypt(plaintext string) (string, error) {
	key, err := getOrCreateKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %w", err)
	}

	hash := sha256.Sum256(key)
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded ciphertext produced by Encrypt() using AES-256-GCM.
func Decrypt(encoded string) (string, error) {
	key, err := getOrCreateKey()
	if err != nil {
		return "", fmt.Errorf("failed to get decryption key: %w", err)
	}

	hash := sha256.Sum256(key)
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	if len(data) < aesGCM.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:aesGCM.NonceSize()], data[aesGCM.NonceSize():]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func tryDecrypt(token string) string {
	if token == "" {
		return token
	}
	decrypted, err := Decrypt(token)
	if err != nil {
		// Not encrypted or failed to decrypt, use original
		return token
	}
	return decrypted
}