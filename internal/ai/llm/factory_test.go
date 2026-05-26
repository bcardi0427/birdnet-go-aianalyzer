package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tphakala/birdnet-go/internal/conf"
)

func TestNewProvider_SupportedProvidersReturnProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
	}{
		{name: "gemini", provider: ProviderGemini},
		{name: "openai", provider: ProviderOpenAI},
		{name: "openrouter", provider: ProviderOpenRouter},
		{name: "openai-compatible", provider: ProviderOpenAICompatible},
		{name: "ollama", provider: ProviderOllama},
		{name: "anthropic", provider: ProviderAnthropic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			provider, err := NewProvider(conf.AISettings{Provider: tt.provider}, "test-key", nil)
			require.NoError(t, err)
			require.NotNil(t, provider)
		})
	}
}

func TestNewProvider_EmptyProviderDefaultsToGemini(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(conf.AISettings{Provider: ""}, "", nil)
	require.NoError(t, err)
	require.NotNil(t, provider)
}

func TestNewProvider_NormalizesProviderID(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(conf.AISettings{Provider: "  OpenAI-Compatible  "}, "", nil)
	require.NoError(t, err)
	require.NotNil(t, provider)
}

func TestNewProvider_RejectsUnsupportedProvider(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(conf.AISettings{Provider: "some-random-provider"}, "", nil)
	require.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "unsupported AI provider")
}

func TestOpenAICompatibleProvider_GenerateAndListModels(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/chat/completions":
			assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
			var req openAIChatRequest
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, "custom-model", req.Model)
			assert.Len(t, req.Messages, 2)
			_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"hello birds"}}]}`))
		case "/models":
			_, _ = w.Write([]byte(`{"data":[{"id":"custom-model","name":"Custom Model","description":"test model"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	provider, err := NewProvider(conf.AISettings{Provider: ProviderOpenAICompatible, BaseURL: server.URL}, "test-key", nil)
	require.NoError(t, err)

	response, err := provider.Generate(context.Background(), GenerateRequest{SystemPrompt: "sys", Prompt: "prompt", Model: "custom-model"})
	require.NoError(t, err)
	assert.Equal(t, "hello birds", response.Text)

	models, err := provider.ListModels(context.Background())
	require.NoError(t, err)
	require.Len(t, models, 1)
	assert.Equal(t, "custom-model", models[0].ID)
	assert.Equal(t, "Custom Model", models[0].DisplayName)
}

func TestAnthropicProvider_GenerateAndListModels(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/messages", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("Anthropic-Version"))
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"anthropic response"}]}`))
	}))
	t.Cleanup(server.Close)

	provider := &anthropicProvider{apiKey: "test-key", baseURL: server.URL, client: server.Client()}
	response, err := provider.Generate(context.Background(), GenerateRequest{SystemPrompt: "sys", Prompt: "prompt", Model: "claude-test"})
	require.NoError(t, err)
	assert.Equal(t, "anthropic response", response.Text)

	models, err := provider.ListModels(context.Background())
	require.NoError(t, err)
	require.Len(t, models, 1)
	assert.Equal(t, DefaultAnthropicModel, models[0].ID)
}
