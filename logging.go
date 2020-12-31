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
	Start time.Time    `json:"start"`
	Url   string       `json:"url"`
	Bot   string       `json:"bot"`
	Game  []JSONStatus `json:game`
}

// NewFileLogger creates a FileLogger with a specified client configuration
func NewFileLogger(config Config) (FileLogger, error) {
	var filename string
	startTime := time.Now()
	for i := 0; ; i++ {
		filename = fmt.Sprintf("%d-%s-%d.json", int64(startTime.Unix()), config.clientName, i)
		_, err := os.Stat(filename)
		if os.IsNotExist(err) {
			break
		} else if err != nil {
			return FileLogger{}, err
		}
	}
	fileLogger := FileLogger{filename: filename, Start: startTime, Url: config.gameURL, Bot: config.clientName, Game: []JSONStatus{}}
	return fileLogger, nil
}

// StoreStatus adds a JSON status to the log data
func (fileLogger *FileLogger) StoreStatus(jsonStatus *JSONStatus) {
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
