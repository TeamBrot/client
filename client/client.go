package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

// Client represents a handler that decides what the specific player should do next
type Client interface {
	GetAction(status *Status, calculationTime time.Duration) Action
}

// RunClient runs a spe_ed client using a specified configuration
func RunClient(config Config) {
	errorLogger := NewErrorLogger()
	clientLogger := NewClientLogger(config.ClientName)
	fileLogger, err := NewFileLogger(config)
	if !info {
		log.SetOutput(ioutil.Discard)
	}
	if err != nil {
		errorLogger.Println("could not create file logger:", err)
	}

	gui := &Gui{nil}
	if config.APIKey != "" {
		guiURL := fmt.Sprintf("%s:%d", config.GUIHostname, config.GUIPort)
		gui = StartGui(guiURL, clientLogger)
	}

	clientLogger.Println("connecting to server")
	conn, err := NewConnection(config)
	if err != nil {
		errorLogger.Fatalln("could not establish connection:", err)
	}
	defer conn.Close()

	status, jsonStatus, err := conn.ReadStatus()
	if err != nil {
		errorLogger.Fatalln("error on first status read:", err)
	}
	fileLogger.Store(jsonStatus)
	clientLogger.Println("field dimensions:", status.Width, "x", status.Height)
	clientLogger.Println("number of players:", len(status.Players))

	defer func() {
		if err := fileLogger.Write(jsonStatus.You); err != nil {
			clientLogger.Println("could not log to file:", err)
		}
	}()

	for jsonStatus.Running && jsonStatus.Players[jsonStatus.You].Active {

		clientLogger.Println("turn", status.Turn)
		clientLogger.Println("deadline", jsonStatus.Deadline)
		me := status.Players[status.You]
		clientLogger.Println("position", me.Y, "y", me.X, "x")
		clientLogger.Println("speed", me.Speed)
		calculationTime, err := computeCalculationTime(jsonStatus.Deadline, config, errorLogger)
		if err != nil {
			errorLogger.Fatalln("error receiving time from server", err)
		}
		clientLogger.Println("the scheduled calculation time is", calculationTime)
		start := time.Now()
		action := config.Client.GetAction(status, calculationTime)
		clientLogger.Println("using", action, "as best action")
		processingTime := time.Since(start)
		if processingTime > calculationTime {
			errorLogger.Println("the calculation took longer then it should have. the client might miss the deadline!")
		}
		clientLogger.Println("the calculation took", processingTime)
		err = conn.WriteAction(action)
		if err != nil {
			errorLogger.Fatalln("error sending action:", err)
		}

		status, jsonStatus, err = conn.ReadStatus()
		if err != nil {
			errorLogger.Fatalln("error reading status:", err)
		}
		fileLogger.Store(jsonStatus)

		err = gui.WriteStatus(jsonStatus)
		if err != nil {
			errorLogger.Println("could not write status to gui:", err)
		}

		counter := 0
		for _, player := range jsonStatus.Players {
			if player.Active {
				counter++
			}
		}
		if counter > 1 {
			clientLogger.Println("active players:", counter)
			if !jsonStatus.Players[jsonStatus.You].Active {
				clientLogger.Println("lost")
			}
		} else if counter == 1 {
			if jsonStatus.Players[jsonStatus.You].Active {
				clientLogger.Println("won")
			} else {
				clientLogger.Println("lost")
			}
		} else {
			clientLogger.Println("lost")
		}
	}
	clientLogger.Println("player inactive, disconnecting...")
}
