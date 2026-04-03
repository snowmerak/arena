package gateway

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ProviderType string

const (
	ProviderOpenAI ProviderType = "openai"
	ProviderClaude ProviderType = "claude"
	ProviderGemini ProviderType = "gemini"
)

type ModelConfig struct {
	Name            string       `yaml:"name"`
	Type            ProviderType `yaml:"type"`
	APIKey          string       `yaml:"api_key"`
	ActualModelName string       `yaml:"actual_model_name"`
	BaseURL         string       `yaml:"base_url,omitempty"`
}

type EmbeddingConfig struct {
	BaseURL             string  `yaml:"base_url"`
	Model               string  `yaml:"model"`
	SimilarityThreshold float64 `yaml:"similarity_threshold"`
}

type RedisConfig struct {
	URL string `yaml:"url"`
}

type Config struct {
	Models    []ModelConfig   `yaml:"models"`
	Embedding EmbeddingConfig `yaml:"embedding"`
	Redis     RedisConfig     `yaml:"redis"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
