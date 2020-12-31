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
	GameURL      string `json:"gameURL"`
	TimeURL      string `json:"timeURL"`
	APIKey       string `json:"-"`
	GUIHostname  string `json:"-"`
	GUIPort      int    `json:"-"`
	LogDirectory string `json:"-"`
	ClientName   string `json:"clientName"`
	Client       Client `json:"-"`
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
	config.GameURL = getenvDefault("URL", defaultGameURL)
	config.TimeURL = getenvDefault("TIME_URL", defaultTimeURL)
	config.APIKey = getenvDefault("KEY", "")

	flag.StringVar(&config.ClientName, "client", "combi", "client to run")
	flag.StringVar(&config.LogDirectory, "log", defaultLogDirectory, "directory in which game statistics are stored")
	flag.StringVar(&config.GUIHostname, "guihostname", defaultGuiHostname, "hostname on which the gui server is listening")
	flag.IntVar(&config.GUIPort, "guiport", defaultGuiPort, "port on which the gui server is listening")
	flag.Parse()

	client, err := getClient(config.ClientName)
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
