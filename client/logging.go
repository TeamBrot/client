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
	Start  time.Time    `json:"start"`
	Game   []JSONStatus `json:"game"`
	Config Config       `json:"config"`
}

// NewFileLogger creates a FileLogger with a specified client configuration
func NewFileLogger(config Config) (FileLogger, error) {
	if _, err := os.Stat(config.LogDirectory); os.IsNotExist(err) {
		os.Mkdir(config.LogDirectory, 0755)
	}
	fileLogger := FileLogger{Start: time.Now(), Game: []JSONStatus{}, Config: config}
	return fileLogger, nil
}

// Store adds a JSON status to the log data
func (fileLogger *FileLogger) Store(jsonStatus *JSONStatus) {
	fileLogger.Game = append(fileLogger.Game, *jsonStatus)
}

// Write writes the game data to disk
func (fileLogger FileLogger) Write(id int) error {
	data, err := json.Marshal(fileLogger)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("%s%c%d-%s-%d.json", fileLogger.Config.LogDirectory, os.PathSeparator, fileLogger.Start.Unix(), fileLogger.Config.ClientName, id)
	err = ioutil.WriteFile(filename, data, 0644)
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
	log.SetOutput(os.Stdout)
	return logger
}

// NewErrorLogger returns a new Logger that writes to Stderr and should be used for all warnings and errors
func NewErrorLogger() *log.Logger {
	logger := log.New(os.Stderr, "[error] ", log.Lmsgprefix|log.LstdFlags)
	return logger
}
