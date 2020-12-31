package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// FileLogger represents all information needed to log a single game to a JSON-encoded file
type FileLogger struct {
	filename string
	Start    time.Time    `json:"start"`
	Game     []JSONStatus `json:"game"`
	Config   Config       `json:"config"`
}

// NewFileLogger creates a FileLogger with a specified client configuration
func NewFileLogger(config Config) (FileLogger, error) {
	if _, err := os.Stat(config.LogDirectory); os.IsNotExist(err) {
		os.Mkdir(config.LogDirectory, 0755)
	}
	var filename string
	startTime := time.Now()
	for i := 0; ; i++ {
		filename = fmt.Sprintf("%s%c%d-%s-%d.json", config.LogDirectory, os.PathSeparator, int64(startTime.Unix()), config.ClientName, i)
		_, err := os.Stat(filename)
		if os.IsNotExist(err) {
			break
		} else if err != nil {
			return FileLogger{}, err
		}
	}
	fileLogger := FileLogger{filename: filename, Start: startTime, Game: []JSONStatus{}, Config: config}
	return fileLogger, nil
}

// Store adds a JSON status to the log data
func (fileLogger *FileLogger) Store(jsonStatus *JSONStatus) {
	fileLogger.Game = append(fileLogger.Game, *jsonStatus)
}

// Write writes the game data to disk
func (fileLogger FileLogger) Write() error {
	data, err := json.Marshal(fileLogger)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileLogger.filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// NewClientLogger creates a logger that is used by all generic client code
func NewClientLogger(clientName string) *log.Logger {
	logger := log.New(os.Stdout, "[client] ", log.Lmsgprefix|log.LstdFlags)
	logger.Println("using client", clientName)
	log.SetPrefix(fmt.Sprintf("[%s] ", clientName))
	log.SetFlags(log.Lmsgprefix | log.LstdFlags)
	return logger
}
