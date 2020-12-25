package main

import (
	"errors"
	"fmt"
	"os"
)

// Config represents a server and client configuration
type Config struct {
	gameURL string
	timeURL string
	apiKey  string
	client  Client
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
	case "mcts":
		client = MctsClient{}
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
	config.gameURL = getenvDefault("URL", "ws://localhost:8080/spe_ed")
	config.timeURL = getenvDefault("TIME_URL", "http://localhost:8080/spe_ed_time")
	config.apiKey = getenvDefault(os.Getenv("KEY"), "")
	client, err := getClient()
	config.client = client
	return config, err
}

// GetWSURL builds the websocket url using the server url and the api key
func (c *Config) GetWSURL() string {
	if c.apiKey == "" {
		return c.gameURL
	}
	return fmt.Sprintf("%s?key=%s", c.gameURL, c.apiKey)
}
