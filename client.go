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

func setupLogging() *log.Logger {
	logger := log.New(os.Stdout, "[client] ", log.Lmsgprefix|log.LstdFlags)
	logger.Println("using client", os.Args[1])
	log.SetPrefix(fmt.Sprintf("[%s] ", os.Args[1]))
	log.SetFlags(log.Lmsgprefix | log.LstdFlags)
	return logger
}

//gets the current server Time via the specified API
func getTime(url string, target interface{}) error {
	httpClient := &http.Client{Timeout: 500 * time.Millisecond}
	r, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

//Sends Signals to the Client after a specified amount of time has passed
func computeCalculationTime(deadline time.Time, config Config) (time.Duration, error) {
	var serverTime ServerTime
	err := getTime(config.TimeURL, &serverTime)
	if err != nil {
		log.Fatalln("couldn't reach timing api")
		return time.Duration(5 * time.Second), err
	}
	calculationTime := deadline.Sub(serverTime.Time)
	calculationTime = time.Duration((calculationTime.Milliseconds() - int64(serverTime.Milliseconds) - 150) * 1000000)
	log.Println("The scheduled calculation Time is :", calculationTime)
	return calculationTime, nil
}

func main() {

	config, err := GetConfig()
	if err != nil {
		fmt.Println("could not get configuration:", err)
		return
	}
	clientLogger := setupLogging()

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

	status, err := conn.ReadStatus()
	if err != nil {
		clientLogger.Fatalln("error on first status read:", err)
	}
	clientLogger.Println("field dimensions:", status.Width, "x", status.Height)
	clientLogger.Println("number of players:", len(status.Players))

	for status.Running && status.Players[status.You].Active {

		clientLogger.Println("turn", status.Turn)
		clientLogger.Println("deadline", status.Deadline)

		calculationTime, err := computeCalculationTime(status.Deadline, config)
		if err != nil {
			clientLogger.Fatalln("error receiving time from server")
		}
		action := config.Client.GetAction(*status.Players[status.You], status, calculationTime)
		err = conn.WriteAction(action)
		if err != nil {
			clientLogger.Fatalln("error sending action:", err)
		}

		status, err = conn.ReadStatus()
		if err != nil {
			clientLogger.Fatalln("error reading status:", err)
		}

		err = gui.WriteStatus(status)
		if err != nil {
			clientLogger.Println("could not write status to gui:", err)
		}

		counter := 0
		for _, player := range status.Players {
			if player.Active {
				counter++
			}
		}
		if counter > 1 {
			clientLogger.Println("active players:", counter)
			if !status.Players[status.You].Active {
				clientLogger.Println("lost")
			}
		} else if counter == 1 {
			if status.Players[status.You].Active {
				clientLogger.Println("won")
				// open output file
				fo, err := os.OpenFile("logging.txt", os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					panic(err)
				}
				// close fo on exit and check for its returned error
				defer func() {
					if err := fo.Close(); err != nil {
						panic(err)
					}
				}()

				fo.WriteString("WON\n")
			} else {
				clientLogger.Println("lost")
				// open output file
				fo, err := os.OpenFile("logging.txt", os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					panic(err)
				}
				// close fo on exit and check for its returned error
				defer func() {
					if err := fo.Close(); err != nil {
						panic(err)
					}
				}()

				fo.WriteString("lost\n")
			}
		} else {
			clientLogger.Println("lost")
			fo, err := os.OpenFile("logging.txt", os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
			// close fo on exit and check for its returned error
			defer func() {
				if err := fo.Close(); err != nil {
					panic(err)
				}
			}()

			fo.WriteString("lost\n")
		}
	}
	clientLogger.Println("player inactive, disconnecting...")
}
