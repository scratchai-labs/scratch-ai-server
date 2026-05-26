package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port          string
	DatabaseURL   string
	DatabasePath  string
	SB3StorageDir string
	DeepSeek      DeepSeekConfig
}

type DeepSeekConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

func FromEnv() Config {
	return Config{
		Port:          envOrDefault("PORT", "8000"),
		DatabaseURL:   strings.TrimSpace(os.Getenv("DATABASE_URL")),
		DatabasePath:  envOrDefault("SERVER_API_DB_PATH", defaultDatabasePath()),
		SB3StorageDir: envOrDefault("SB3_STORAGE_DIR", defaultSB3StorageDir()),
		DeepSeek: DeepSeekConfig{
			BaseURL: strings.TrimRight(envOrDefault("DEEPSEEK_BASE_URL", "https://api.deepseek.com"), "/"),
			APIKey:  strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
			Model:   envOrDefault("DEEPSEEK_MODEL", "deepseek-v4-flash"),
			Timeout: deepSeekTimeout(),
		},
	}
}

func defaultSB3StorageDir() string {
	return filepath.Join(os.TempDir(), "scratch-ai-server-sb3")
}

func defaultDatabasePath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("scratch-ai-server-%d.sqlite3", time.Now().UTC().UnixNano()))
}

func (c DeepSeekConfig) Enabled() bool {
	return c.APIKey != ""
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func deepSeekTimeout() time.Duration {
	raw := strings.TrimSpace(os.Getenv("DEEPSEEK_TIMEOUT_SECONDS"))
	if raw == "" {
		return 8 * time.Second
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return 8 * time.Second
	}
	return time.Duration(seconds) * time.Second
}
