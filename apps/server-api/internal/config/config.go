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
	Mode                    string
	Port                    string
	DatabaseURL             string
	DatabasePath            string
	SB3StorageDir           string
	SB3StorageDirConfigured bool
	CORSAllowedOrigins      []string
	DeepSeek                DeepSeekConfig
}

type DeepSeekConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

func FromEnv() Config {
	sb3StorageDir, hasSB3StorageDir := envValue("SB3_STORAGE_DIR")

	return Config{
		Mode:                    envOrDefault("GIN_MODE", "debug"),
		Port:                    envOrDefault("PORT", "8000"),
		DatabaseURL:             strings.TrimSpace(os.Getenv("DATABASE_URL")),
		DatabasePath:            envOrDefault("SERVER_API_DB_PATH", defaultDatabasePath()),
		SB3StorageDir:           valueOrFallback(sb3StorageDir, defaultSB3StorageDir()),
		SB3StorageDirConfigured: hasSB3StorageDir,
		CORSAllowedOrigins:      parseCSVEnv("CORS_ALLOWED_ORIGINS"),
		DeepSeek: DeepSeekConfig{
			BaseURL: strings.TrimRight(envOrDefault("DEEPSEEK_BASE_URL", "https://api.deepseek.com"), "/"),
			APIKey:  strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
			Model:   envOrDefault("DEEPSEEK_MODEL", "deepseek-v4-flash"),
			Timeout: deepSeekTimeout(),
		},
	}
}

func (c Config) ValidateForRuntime() error {
	if !strings.EqualFold(strings.TrimSpace(c.Mode), "release") {
		return nil
	}

	missing := make([]string, 0, 3)
	if strings.TrimSpace(c.DatabaseURL) == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if !c.SB3StorageDirConfigured {
		missing = append(missing, "SB3_STORAGE_DIR")
	}
	if len(c.CORSAllowedOrigins) == 0 || containsWildcardOrigin(c.CORSAllowedOrigins) {
		missing = append(missing, "CORS_ALLOWED_ORIGINS")
	}
	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("release mode requires %s", strings.Join(missing, ", "))
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
	return valueOrFallback(strings.TrimSpace(os.Getenv(key)), fallback)
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

func envValue(key string) (string, bool) {
	value := strings.TrimSpace(os.Getenv(key))
	return value, value != ""
}

func parseCSVEnv(key string) []string {
	raw, ok := envValue(key)
	if !ok {
		return nil
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func valueOrFallback(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func containsWildcardOrigin(origins []string) bool {
	for _, origin := range origins {
		if strings.TrimSpace(origin) == "*" {
			return true
		}
	}
	return false
}
