package config

import (
	"errors"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Port                 string
	OpenAIBaseURL        string
	OpenAIKey            string
	DatabaseURL          string
	OpenAIModel          string
	OpenAIEmbeddingModel string
	SessionHistorySize   int
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetDefault("PORT", "8080")
	v.SetDefault("OPENAI_MODEL", "gpt-4o-mini")
	v.SetDefault("OPENAI_EMBEDDING_MODEL", "text-embedding-3-large")
	v.SetDefault("SESSION_HISTORY_SIZE", 10)

	v.AutomaticEnv()

	// Support config.yaml plus .env files for local development.
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	_ = v.ReadInConfig()

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig()

	cfg := &Config{
		Port:                 v.GetString("PORT"),
		OpenAIBaseURL:        v.GetString("OPENAI_BASE_URL"),
		OpenAIKey:            v.GetString("OPENAI_API_KEY"),
		DatabaseURL:          v.GetString("DATABASE_URL"),
		OpenAIModel:          v.GetString("OPENAI_MODEL"),
		OpenAIEmbeddingModel: v.GetString("OPENAI_EMBEDDING_MODEL"),
		SessionHistorySize:   v.GetInt("SESSION_HISTORY_SIZE"),
	}

	if cfg.OpenAIKey == "" {
		openAIKey := os.Getenv("OPENAI_API_KEY")
		if openAIKey != "" {
			cfg.OpenAIKey = openAIKey
		}
		return nil, ErrMissingOpenAIKey
	}
	if cfg.DatabaseURL == "" {
		return nil, ErrMissingDatabaseURL
	}

	return cfg, nil
}

var (
	ErrMissingOpenAIKey   = errors.New("OPENAI_API_KEY 环境变量未配置")
	ErrMissingDatabaseURL = errors.New("DATABASE_URL 环境变量未配置")
)
