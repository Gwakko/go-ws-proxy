package configs

import (
	"os"
	"strconv"
)

type Config struct {
	Port           int
	AllowedOrigins string
	CommandTimeout int // seconds
	AllowlistPath  string
}

func Load() *Config {
	return &Config{
		Port:           getEnvInt("PORT", 8080),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
		CommandTimeout: getEnvInt("COMMAND_TIMEOUT", 30),
		AllowlistPath:  getEnv("ALLOWLIST_PATH", "configs/allowlist.json"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
