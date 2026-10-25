package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBURL        string
	Port         int
	Username     string
	Password     string
	CookieSecure bool
}

func Load() (*Config, error) {
	dburl := os.Getenv("DATABASE_URL")
	if dburl == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")

	}
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT %q: %w", p, err)
		}
		port = n
	}

	return &Config{
		DBURL:        dburl,
		Port:         port,
		Username:     os.Getenv("USERNAME"),
		Password:     os.Getenv("PASSWORD"),
		CookieSecure: os.Getenv("COOKIE_SECURE") != "false",
	}, nil
}
