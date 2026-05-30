package conf

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	configEncryptionPrefix  = "enc:v1:"
	configEncryptionKeyEnv  = "BIRDNET_CONFIG_ENCRYPTION_KEY"
	configEncryptionKeyFile = "config.encryption.key"
)

func isEncryptedValue(v string) bool { return strings.HasPrefix(v, configEncryptionPrefix) }

func encryptValue(key []byte, plain string) (string, error) {
	if strings.TrimSpace(plain) == "" || isEncryptedValue(plain) {
		return plain, nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM block cipher: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to read random nonce: %w", err)
	}
	ct := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return configEncryptionPrefix + base64.StdEncoding.EncodeToString(ct), nil
}

func decryptValue(key []byte, value string) (string, error) {
	if !isEncryptedValue(value) {
		return value, nil
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(value, configEncryptionPrefix))
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM block cipher: %w", err)
	}
	if len(raw) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ct := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt ciphertext: %w", err)
	}
	return string(pt), nil
}

func configEncryptionKey() ([]byte, error) {
	if env := strings.TrimSpace(os.Getenv(configEncryptionKeyEnv)); env != "" {
		return decodeKeyString(env)
	}
	paths, err := GetDefaultConfigPaths()
	if err != nil {
		return nil, fmt.Errorf("config path unavailable: %w", err)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("config path list is empty")
	}
	kp := filepath.Join(paths[0], configEncryptionKeyFile)
	b, err := os.ReadFile(kp)
	if err == nil {
		key, decErr := decodeKeyString(strings.TrimSpace(string(b)))
		if decErr != nil {
			return nil, fmt.Errorf("failed to decode encryption key file: %w", decErr)
		}
		return key, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read encryption key file: %w", err)
	}
	key := make([]byte, 32)
	if _, err = rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	if err = os.MkdirAll(filepath.Dir(kp), 0o750); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	if err = os.WriteFile(kp, []byte(hex.EncodeToString(key)), 0o600); err != nil {
		return nil, fmt.Errorf("failed to write encryption key file: %w", err)
	}
	return key, nil
}

func decodeKeyString(s string) ([]byte, error) {
	if b, err := hex.DecodeString(s); err == nil && len(b) == 32 {
		return b, nil
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid config encryption key: failed to decode base64: %w", err)
	}
	if len(b) != 32 {
		return nil, fmt.Errorf("invalid config encryption key length: expected 32 bytes, got %d", len(b))
	}
	return b, nil
}
