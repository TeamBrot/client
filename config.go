package main

import (
	"fmt"
	"log"
	"os"
)

type Config struct {
	gameURL string
	timeURL string
	apiKey  string
}

func getenvDefault(key string, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	return value
}

func GetConfig(logger *log.Logger) Config {
	var config Config
	config.gameURL = getenvDefault("URL", "ws://localhost:8080/spe_ed")
	config.timeURL = getenvDefault("TIME_URL", "http://localhost:8080/spe_ed_time")
	config.apiKey = getenvDefault(os.Getenv("KEY"), "")
	return config
}

func (c *Config) GetWSURL() string {
	if c.apiKey == "" {
		return c.gameURL
	}
	return fmt.Sprintf("%s?key=%s", c.gameURL, c.apiKey)
}
