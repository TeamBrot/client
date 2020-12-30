package main

import (
	"errors"
	"fmt"
	"os"
)

// Config represents a server and client configuration
type Config struct {
	GameURL string
	TimeURL string
	APIKey  string
	Client  Client
}

func getenvDefault(key string, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	return value
}

func getClient() (Client, error) {
	if len(os.Args) <= 1 {
		return nil, errors.New("no client given")
	}
	var client Client
	switch os.Args[1] {
	case "minimax":
		client = MinimaxClient{}
		break
	case "smart":
		client = SmartClient{}
		break
	case "speku":
		client = SpekuClient{}
		break
	default:
		return nil, fmt.Errorf("invalid client name: %s", os.Args[1])
	}
	return client, nil
}

// GetConfig creates a config from the environment variables
func GetConfig() (Config, error) {
	var config Config
	config.GameURL = getenvDefault("URL", defaultGameURL)
	config.TimeURL = getenvDefault("TIME_URL", defaultTimeURL)
	config.APIKey = getenvDefault("KEY", "")
	client, err := getClient()
	config.Client = client
	return config, err
}

// GetWSURL builds the websocket url using the server url and the api key
func (c *Config) GetWSURL() string {
	if c.APIKey == "" {
		return c.GameURL
	}
	return fmt.Sprintf("%s?key=%s", c.GameURL, c.APIKey)
}
