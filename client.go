package main

import (
	"fmt"
	"time"
)

// Client represents a handler that decides what the specific player should do next
type Client interface {
	GetAction(player Player, status *Status, calculationTime time.Duration) Action
}

// RunClient runs a spe_ed client using a specified configuration
func RunClient(config Config) {
	clientLogger := NewClientLogger(config.clientName)
	fileLogger, err := NewFileLogger(config)
	if err != nil {
		clientLogger.Println("could not create file logger:", err)
	}
	defer func() {
		if err := fileLogger.Write(); err != nil {
			clientLogger.Println("could not log to file:", err)
		}
	}()

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
			}
		} else if counter == 1 {
			if JSONStatus.Players[JSONStatus.You].Active {
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
