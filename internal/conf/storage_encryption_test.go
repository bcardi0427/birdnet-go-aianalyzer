package conf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPersistMigration_EncryptsSecrets(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// Set env encryption key to ensure repeatable key and avoid generating a persistent file
	t.Setenv(configEncryptionKeyEnv, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	// Initialize settings with plaintext keys
	s := createMinimalValidSettings()
	s.AI.Enabled = true
	s.AI.Provider = "openai"
	s.AI.APIKey = "super-secret-openai-api-key"
	s.Security.SessionSecret = "plaintext-session-secret"

	// Mock viper ConfigFile
	viper.Reset()
	viper.SetConfigFile(configFile)
	// Write dummy config to configFile so we can read it later
	err := SaveYAMLConfig(configFile, s)
	require.NoError(t, err)

	// Call persistMigration
	persistMigration(s, "test-migration")

	// Verify that the original s was NOT modified (in-memory settings are still plaintext)
	assert.Equal(t, "super-secret-openai-api-key", s.AI.APIKey)
	assert.Equal(t, "plaintext-session-secret", s.Security.SessionSecret)

	// Read the config file from disk and unmarshal into a raw map to verify encryption
	data, err := os.ReadFile(configFile)
	require.NoError(t, err)

	var raw map[string]any
	err = yaml.Unmarshal(data, &raw)
	require.NoError(t, err)

	// Verify that AI.APIKey and Security.SessionSecret are encrypted (prefixed with enc:v1:)
	aiSection, ok := raw["ai"].(map[string]any)
	require.True(t, ok)
	apiKeyVal, ok := aiSection["apikey"].(string)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(apiKeyVal, configEncryptionPrefix), "API key should be encrypted, got: %s", apiKeyVal)

	securitySection, ok := raw["security"].(map[string]any)
	require.True(t, ok)
	sessionSecretVal, ok := securitySection["sessionsecret"].(string)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(sessionSecretVal, configEncryptionPrefix), "Session secret should be encrypted, got: %s", sessionSecretVal)
}

func TestEnsureSessionSecret_EncryptsSecrets(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// Set env encryption key to ensure repeatable key
	t.Setenv(configEncryptionKeyEnv, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	// Initialize settings with empty SessionSecret
	s := createMinimalValidSettings()
	s.AI.Enabled = true
	s.AI.Provider = "openai"
	s.AI.APIKey = "another-secret-api-key"
	s.Security.SessionSecret = "" // Empty to trigger generation

	// Mock viper ConfigFile
	viper.Reset()
	viper.SetConfigFile(configFile)
	err := SaveYAMLConfig(configFile, s)
	require.NoError(t, err)

	// Call ensureSessionSecret
	err = ensureSessionSecret(s)
	require.NoError(t, err)

	// Verify that in-memory settings now has a generated session secret in plaintext
	assert.NotEmpty(t, s.Security.SessionSecret)
	assert.False(t, strings.HasPrefix(s.Security.SessionSecret, configEncryptionPrefix))
	assert.Equal(t, "another-secret-api-key", s.AI.APIKey)

	// Read the config file from disk and unmarshal into a raw map to verify encryption
	data, err := os.ReadFile(configFile)
	require.NoError(t, err)

	var raw map[string]any
	err = yaml.Unmarshal(data, &raw)
	require.NoError(t, err)

	// Verify that AI.APIKey and Security.SessionSecret are encrypted (prefixed with enc:v1:)
	aiSection, ok := raw["ai"].(map[string]any)
	require.True(t, ok)
	apiKeyVal, ok := aiSection["apikey"].(string)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(apiKeyVal, configEncryptionPrefix), "API key should be encrypted, got: %s", apiKeyVal)

	securitySection, ok := raw["security"].(map[string]any)
	require.True(t, ok)
	sessionSecretVal, ok := securitySection["sessionsecret"].(string)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(sessionSecretVal, configEncryptionPrefix), "Session secret should be encrypted, got: %s", sessionSecretVal)
}
