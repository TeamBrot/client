package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Client represents a handler that decides what the specific player should do next
type Client interface {
	GetAction(player Player, status *Status, calculationTime time.Duration) Action
}

func newClientLogger(clientName string) *log.Logger {
	logger := log.New(os.Stdout, "[client] ", log.Lmsgprefix|log.LstdFlags)
	logger.Println("using client", clientName)
	log.SetPrefix(fmt.Sprintf("[%s] ", clientName))
	log.SetFlags(log.Lmsgprefix | log.LstdFlags)
	return logger
}

func newFileLogger(filename string) (*log.Logger, func(), error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}
	logger := log.New(file, "", log.LstdFlags)
	return logger, func() { file.Close() }, nil
}
func RunClient(config Config) {
	clientLogger := newClientLogger(config.clientName)
	fileLogger, closeFunc, err := newFileLogger(defaultLogFile)
	if err != nil {
		clientLogger.Println("could not create fileLogger:", err)
	}
	defer closeFunc()

	gui := &Gui{nil}
	if config.apiKey != "" {
		guiUrl := fmt.Sprintf("%s:%d", config.guiHostname, config.guiPort)
		gui = StartGui(guiUrl, clientLogger)
	}

	clientLogger.Println("connecting to server")
	conn, err := NewConnection(config)
	if err != nil {
		clientLogger.Fatalln("could not establish connection:", err)
	}
	defer conn.Close()

	status, JSONStatus, err := conn.ReadStatus()
	if err != nil {
		clientLogger.Fatalln("error on first status read:", err)
	}
	clientLogger.Println("field dimensions:", status.Width, "x", status.Height)
	clientLogger.Println("number of players:", len(status.Players))

	for JSONStatus.Running && JSONStatus.Players[JSONStatus.You].Active {

		clientLogger.Println("turn", status.Turn)
		clientLogger.Println("deadline", JSONStatus.Deadline)
		me := status.Players[status.You]
		clientLogger.Println("Position ", me.Y, "y", me.X, "x")
		clientLogger.Println("Speed", me.Speed)
		calculationTime, err := computeCalculationTime(JSONStatus.Deadline, config)
		if err != nil {
			clientLogger.Fatalln("error receiving time from server")
		}

		action := config.client.GetAction(*status.Players[status.You], status, calculationTime)
		err = conn.WriteAction(action)
		if err != nil {
			clientLogger.Fatalln("error sending action:", err)
		}

		status, JSONStatus, err = conn.ReadStatus()
		if err != nil {
			clientLogger.Fatalln("error reading status:", err)
		}

		err = gui.WriteStatus(JSONStatus)
		if err != nil {
			clientLogger.Println("could not write status to gui:", err)
		}

		counter := 0
		for _, player := range JSONStatus.Players {
			if player.Active {
				counter++
			}
		}
		if counter > 1 {
			clientLogger.Println("active players:", counter)
			if !JSONStatus.Players[JSONStatus.You].Active {
				clientLogger.Println("lost")
				fileLogger.Println("At Beginnig there were", len(JSONStatus.Players), "Players")
				fileLogger.Println("Now are", len(status.Players), "Players still Active")
				fileLogger.Println("The Game lasted", status.Turn, "Turns")
				fileLogger.Println("The field hat the dimensions", status.Width, "x", status.Height)
				fileLogger.Println("lost")
			}
		} else if counter == 1 {
			if JSONStatus.Players[JSONStatus.You].Active {
				clientLogger.Println("won")
				fileLogger.Println("At Beginnig there were", len(JSONStatus.Players), "Players")
				fileLogger.Println("Now are", len(status.Players), "Players still Active")
				fileLogger.Println("The Game lasted", status.Turn, "Turns")
				fileLogger.Println("The field hat the dimensions", status.Width, "x", status.Height)
				fileLogger.Println("won")
			} else {
				clientLogger.Println("lost")
				fileLogger.Println("At Beginnig there were", len(JSONStatus.Players), "Players")
				fileLogger.Println("Now are", len(status.Players), "Players still Active")
				fileLogger.Println("The Game lasted", status.Turn, "Turns")
				fileLogger.Println("The field hat the dimensions", status.Width, "x", status.Height)
				fileLogger.Println("lost")
			}
		} else {
			clientLogger.Println("lost")
			fileLogger.Println("At Beginnig there were", len(JSONStatus.Players), "Players")
			fileLogger.Println("Now are", len(status.Players), "Players still Active")
			fileLogger.Println("The Game lasted", status.Turn, "Turns")
			fileLogger.Println("The field hat the dimensions", status.Width, "x", status.Height)
			fileLogger.Println("lost")
		}
	}
	clientLogger.Println("player inactive, disconnecting...")
}
