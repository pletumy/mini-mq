package config

import "os"

type Config struct {
	DBUrl      string
	ServerPort string
}

func Load() Config {
	return Config{
		DBUrl:      getEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/mini_mq?sslmode=disable"),
		ServerPort: getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
