package conf

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type SecretStore struct {
	Secrets map[string]string `yaml:"secrets"`
}

func getSecretFields(s *Settings) map[string]*string {
	m := map[string]*string{
		"ai.api_key":                            &s.AI.APIKey,
		"ai.gemini.api_key":                     &s.AI.Gemini.APIKey,
		"ai.openai.api_key":                     &s.AI.OpenAI.APIKey,
		"ai.openrouter.api_key":                 &s.AI.OpenRouter.APIKey,
		"ai.openaicompatible.api_key":           &s.AI.OpenAICompatible.APIKey,
		"ai.ollama.api_key":                     &s.AI.Ollama.APIKey,
		"ai.anthropic.api_key":                  &s.AI.Anthropic.APIKey,
		"realtime.ebird.api_key":                &s.Realtime.EBird.APIKey,
		"realtime.weather.openweather.api_key":  &s.Realtime.Weather.OpenWeather.APIKey,
		"realtime.weather.wunderground.api_key": &s.Realtime.Weather.Wunderground.APIKey,
		"realtime.mqtt.password":                &s.Realtime.MQTT.Password,
		"output.mysql.password":                 &s.Output.MySQL.Password,
		"security.session_secret":               &s.Security.SessionSecret,
		"security.basicauth.password":           &s.Security.BasicAuth.Password,
		"security.googleauth.client_secret":     &s.Security.GoogleAuth.ClientSecret,
		"security.githubauth.client_secret":     &s.Security.GithubAuth.ClientSecret,
		"security.microsoftauth.client_secret":  &s.Security.MicrosoftAuth.ClientSecret,
		"backup.encryption_key":                 &s.Backup.EncryptionKey,
	}
	for i := range s.Security.OAuthProviders {
		m[fmt.Sprintf("security.oauth_providers.%d.client_secret", i)] = &s.Security.OAuthProviders[i].ClientSecret
	}
	for i := range s.Notification.Push.Providers {
		for j := range s.Notification.Push.Providers[i].Endpoints {
			a := &s.Notification.Push.Providers[i].Endpoints[j].Auth
			m[fmt.Sprintf("notification.push.providers.%d.endpoints.%d.auth.token", i, j)] = &a.Token
			m[fmt.Sprintf("notification.push.providers.%d.endpoints.%d.auth.pass", i, j)] = &a.Pass
			m[fmt.Sprintf("notification.push.providers.%d.endpoints.%d.auth.value", i, j)] = &a.Value
		}
	}
	return m
}

func getSecretsPath() (string, error) {
	configPath, err := FindConfigFile()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(configPath), "secrets.yaml"), nil
}

func loadSecrets(s *Settings) error {
	path, err := getSecretsPath()
	if err != nil {
		return err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var store SecretStore
	if err := yaml.Unmarshal(b, &store); err != nil {
		return err
	}

	key, err := configEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to get config encryption key: %w", err)
	}

	fields := getSecretFields(s)
	for k, encVal := range store.Secrets {
		if ptr, ok := fields[k]; ok {
			decVal, err := decryptValue(key, encVal)
			if err != nil {
				return fmt.Errorf("decryption failed for %s: %w", k, err)
			}
			*ptr = decVal
		}
	}
	return nil
}

func saveSecrets(s *Settings) error {
	path, err := getSecretsPath()
	if err != nil {
		return err
	}

	key, err := configEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to get config encryption key: %w", err)
	}

	store := SecretStore{
		Secrets: make(map[string]string),
	}

	fields := getSecretFields(s)
	for k, ptr := range fields {
		if *ptr != "" {
			encVal, err := encryptValue(key, *ptr)
			if err != nil {
				return fmt.Errorf("encryption failed for %s: %w", k, err)
			}
			store.Secrets[k] = encVal
			// Zero out so it doesn't get saved to config.yaml
			*ptr = ""
		}
	}

	b, err := yaml.Marshal(&store)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}
