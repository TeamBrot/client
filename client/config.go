package main

import (
	"flag"
	"fmt"
	"os"
)

const defaultGameURL = "ws://localhost:8080/spe_ed"
const defaultTimeURL = "http://localhost:8080/spe_ed_time"
const defaultLogFile = "logging.txt"
const defaultGuiHostname = "0.0.0.0"
const defaultGuiPort = 8081
const defaultLogDirectory = "log"

// Config represents a server and client configuration
type Config struct {
	gameURL      string
	timeURL      string
	apiKey       string
	guiHostname  string
	guiPort      int
	logDirectory string
	clientName   string
	client       Client
}

func getenvDefault(key string, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	return value
}

func getClient(name string) (Client, error) {
	var client Client
	switch name {
	case "minimax":
		client = MinimaxClient{}
		break
	case "smart":
		client = SmartClient{}
		break
	case "combi":
		client = CombiClient{}
		break
	case "rollouts":
		client = RolloutClient{}
		break
	case "probability":
		client = ProbabilityClient{}
		break
	default:
		return nil, fmt.Errorf("invalid client name: %s", name)
	}
	return client, nil
}

// GetConfig creates a config from the environment variables
func GetConfig() (Config, error) {
	var config Config
	config.gameURL = getenvDefault("URL", defaultGameURL)
	config.timeURL = getenvDefault("TIME_URL", defaultTimeURL)
	config.apiKey = getenvDefault("KEY", "")

	flag.StringVar(&config.clientName, "client", "combi", "client to run")
	flag.StringVar(&config.logDirectory, "log", defaultLogDirectory, "directory in which game statistics are stored")
	flag.StringVar(&config.guiHostname, "guihostname", defaultGuiHostname, "hostname on which the gui server is listening")
	flag.IntVar(&config.guiPort, "guiport", defaultGuiPort, "port on which the gui server is listening")
	flag.Parse()

	client, err := getClient(config.clientName)
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
