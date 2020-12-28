package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type ServerTime struct {
	Time         time.Time `json:"time"`
	Milliseconds int       `json:"milliseconds"`
}

// Client represents a handler that decides what the specific player should do next
type Client interface {
	GetAction(player Player, status *Status, calculationTime time.Duration) Action
}

func newClientLogger() *log.Logger {
	logger := log.New(os.Stdout, "[client] ", log.Lmsgprefix|log.LstdFlags)
	logger.Println("using client", os.Args[1])
	log.SetPrefix(fmt.Sprintf("[%s] ", os.Args[1]))
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

var httpClient http.Client = http.Client{Timeout: 500 * time.Millisecond}

//gets the current server Time via the specified API
func getTime(url string) (ServerTime, error) {
	var time ServerTime
	r, err := httpClient.Get(url)
	if err != nil {
		return time, err
	}
	defer r.Body.Close()

	json.NewDecoder(r.Body).Decode(&time)
	return time, nil
}

//Sends Signals to the Client after a specified amount of time has passed
func computeCalculationTime(deadline time.Time, config Config) (time.Duration, error) {
	serverTime, err := getTime(config.TimeURL)
	if err != nil {
		log.Println("couldn't reach timing api, using 5s timeout")
		return time.Duration(5 * time.Second), err
	}
	calculationTime := deadline.Sub(serverTime.Time)
	calculationTime = time.Duration((calculationTime.Milliseconds() - int64(serverTime.Milliseconds) - 150) * 1000000)
	log.Println("the scheduled calculation Time is", calculationTime)
	return calculationTime, nil
}

func main() {

	config, err := GetConfig()
	if err != nil {
		fmt.Println("could not get configuration:", err)
		return
	}
	clientLogger := newClientLogger()
	fileLogger, closeFunc, err := newFileLogger("logging.txt")
	if err != nil {
		clientLogger.Println("could not create fileLogger:", err)
	}
	defer closeFunc()

	gui := &Gui{nil}
	if config.APIKey != "" {
		gui = StartGui(clientLogger)
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

		calculationTime, err := computeCalculationTime(JSONStatus.Deadline, config)
		if err != nil {
			clientLogger.Fatalln("error receiving time from server")
		}
		action := config.Client.GetAction(*status.Players[status.You], status, calculationTime)
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
				fileLogger.Println("won")
			} else {
				clientLogger.Println("lost")
				fileLogger.Println("lost")
			}
		} else {
			clientLogger.Println("lost")
			fileLogger.Println("lost")
		}
	}
	clientLogger.Println("player inactive, disconnecting...")
}
