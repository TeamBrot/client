package main

import (
	"flag"
	"fmt"
	"os"
)

//Activates more informative logging
const info = true
const defaultGameURL = "ws://localhost:8080/spe_ed"
const defaultTimeURL = "http://localhost:8080/spe_ed_time"
const defaultLogFile = "logging.txt"
const defaultGuiHostname = "0.0.0.0"
const defaultGuiPort = 8081
const defaultLogDirectory = "log"

// If the sum of all probabilities in the specified window is higher then this, minimax can be used
const defaultMinimaxActivationValue = 0.001
const defaultMyStartProbability = 2.0
const defaultFilterValue = 0.69

// Config represents a server and client configuration
type Config struct {
	GameURL                string  `json:"gameURL"`
	TimeURL                string  `json:"timeURL"`
	APIKey                 string  `json:"-"`
	GUIHostname            string  `json:"-"`
	GUIPort                int     `json:"-"`
	LogDirectory           string  `json:"-"`
	ClientName             string  `json:"clientName"`
	Client                 Client  `json:"-"`
	MinimaxActivationValue float64 `json:"minimaxActivationValue"`
	MyStartProbability     float64 `json:"myStartProbability"`
	FilterValue            float64 `json:"filterValue"`
}

func getenvDefault(key string, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	return value
}

func getClient(config Config) (Client, error) {
	var client Client
	switch config.ClientName {
	case "minimax":
		client = MinimaxClient{}
		break
	case "smart":
		client = SmartClient{}
		break
	case "combi":
		client = CombiClient{config.MinimaxActivationValue, config.MyStartProbability, config.FilterValue}
		break
	case "rollouts":
		client = RolloutClient{config.FilterValue}
		break
	case "probability":
		client = ProbabilityClient{config.MyStartProbability}
		break
	default:
		return nil, fmt.Errorf("invalid client name: %s", config.ClientName)
	}
	return client, nil
}

// GetConfig creates a config from the environment variables
func GetConfig() (Config, error) {
	var config Config
	config.GameURL = getenvDefault("URL", defaultGameURL)
	config.TimeURL = getenvDefault("TIME_URL", defaultTimeURL)
	config.APIKey = getenvDefault("KEY", "")
	flag.Float64Var(&config.FilterValue, "filter", defaultFilterValue, "defines the filterValue used in rollouts")
	flag.Float64Var(&config.MinimaxActivationValue, "activation", defaultMinimaxActivationValue, "defines minimaxActivationValue")
	flag.Float64Var(&config.MyStartProbability, "probability", defaultMyStartProbability, "defines myStartProbability")
	flag.StringVar(&config.ClientName, "client", "combi", "client to run")
	flag.StringVar(&config.LogDirectory, "log", defaultLogDirectory, "directory in which game statistics are stored")
	flag.StringVar(&config.GUIHostname, "guihostname", defaultGuiHostname, "hostname on which the gui server is listening")
	flag.IntVar(&config.GUIPort, "guiport", defaultGuiPort, "port on which the gui server is listening")
	flag.Parse()

	client, err := getClient(config)
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
