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

	// Read the config file from disk and unmarshal into a raw map to verify they are cleared/empty in config.yaml
	data, err := os.ReadFile(configFile)
	require.NoError(t, err)

	var raw map[string]any
	err = yaml.Unmarshal(data, &raw)
	require.NoError(t, err)

	aiSection, ok := raw["ai"].(map[string]any)
	require.True(t, ok)
	apiKeyVal, ok := aiSection["apikey"].(string)
	require.True(t, ok)
	assert.Empty(t, apiKeyVal, "API key should be cleared in config.yaml, got: %s", apiKeyVal)

	securitySection, ok := raw["security"].(map[string]any)
	require.True(t, ok)
	sessionSecretVal, ok := securitySection["sessionsecret"].(string)
	require.True(t, ok)
	assert.Empty(t, sessionSecretVal, "Session secret should be cleared in config.yaml, got: %s", sessionSecretVal)

	// Verify they are encrypted in secrets.yaml
	secretsFile := filepath.Join(filepath.Dir(configFile), "secrets.yaml")
	secretsData, err := os.ReadFile(secretsFile)
	require.NoError(t, err)

	var store SecretStore
	err = yaml.Unmarshal(secretsData, &store)
	require.NoError(t, err)

	encAPIKey, ok := store.Secrets["ai.api_key"]
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(encAPIKey, configEncryptionPrefix), "API key in secrets.yaml should be encrypted, got: %s", encAPIKey)

	encSessionSecret, ok := store.Secrets["security.session_secret"]
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(encSessionSecret, configEncryptionPrefix), "Session secret in secrets.yaml should be encrypted, got: %s", encSessionSecret)
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

	// Read the config file from disk and unmarshal into a raw map to verify they are cleared/empty in config.yaml
	data, err := os.ReadFile(configFile)
	require.NoError(t, err)

	var raw map[string]any
	err = yaml.Unmarshal(data, &raw)
	require.NoError(t, err)

	aiSection, ok := raw["ai"].(map[string]any)
	require.True(t, ok)
	apiKeyVal, ok := aiSection["apikey"].(string)
	require.True(t, ok)
	assert.Empty(t, apiKeyVal, "API key should be cleared in config.yaml, got: %s", apiKeyVal)

	securitySection, ok := raw["security"].(map[string]any)
	require.True(t, ok)
	sessionSecretVal, ok := securitySection["sessionsecret"].(string)
	require.True(t, ok)
	assert.Empty(t, sessionSecretVal, "Session secret should be cleared in config.yaml, got: %s", sessionSecretVal)

	// Verify they are encrypted in secrets.yaml
	secretsFile := filepath.Join(filepath.Dir(configFile), "secrets.yaml")
	secretsData, err := os.ReadFile(secretsFile)
	require.NoError(t, err)

	var store SecretStore
	err = yaml.Unmarshal(secretsData, &store)
	require.NoError(t, err)

	encAPIKey, ok := store.Secrets["ai.api_key"]
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(encAPIKey, configEncryptionPrefix), "API key in secrets.yaml should be encrypted, got: %s", encAPIKey)

	encSessionSecret, ok := store.Secrets["security.session_secret"]
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(encSessionSecret, configEncryptionPrefix), "Session secret in secrets.yaml should be encrypted, got: %s", encSessionSecret)
}
