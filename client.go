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
	GetAction(player Player, status *Status, timingChannel <-chan time.Time) Action
}

func setupLogging() *log.Logger {
	logger := log.New(os.Stdout, "[client] ", log.Lmsgprefix|log.LstdFlags)
	logger.Println("using client", os.Args[1])
	log.SetPrefix(fmt.Sprintf("[%s] ", os.Args[1]))
	log.SetFlags(log.Lmsgprefix | log.LstdFlags)
	return logger
}

//gets the current server Time via the specified API
func getTime(url string, target interface{}, httpClient *http.Client) error {
	r, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

//Sends Signals to the Client after a specified amount of time has passed
func calculateTiming(deadline time.Time, serverTime ServerTime, timingChannel chan<- time.Time) {
	calculationTime := deadline.Sub(serverTime.Time)
	calculationTime = time.Duration((calculationTime.Milliseconds() - int64(serverTime.Milliseconds) - 150) * 1000000)
	log.Println("The scheduled calculation Time is :", calculationTime)
	//send First Singal(Calculations that need time to finish are interrupted)
	time.Sleep(time.Duration(0.4 * float64(calculationTime.Nanoseconds())))
	if timingChannel != nil {
		timingChannel <- time.Now()
	} else {
		return
	}
	//send Second Signal (All Calculations are interripted, the results are interpreted)
	time.Sleep(time.Duration(0.4 * float64(calculationTime)))
	if timingChannel != nil {
		timingChannel <- time.Now()
	}
	time.Sleep(time.Duration(0.1 * float64(calculationTime)))
	if timingChannel != nil {
		timingChannel <- time.Now()
	}
	close(timingChannel)
}

func main() {
	config, err := GetConfig()
	if err != nil {
		fmt.Println("could not get configuration:", err)
		return
	}
	clientLogger := setupLogging()
	httpClient := &http.Client{Timeout: 2 * time.Second}

	clientLogger.Println("connecting to server")
	conn, err := NewConnection(config)
	if err != nil {
		clientLogger.Fatalln("could not establish connection:", err)
	}
	defer conn.Close()

	var serverTime ServerTime
	var timingChannel chan time.Time

	status, err := conn.ReadStatus()
	if err != nil {
		clientLogger.Fatalln("error on first status read:", err)
	}
	clientLogger.Println("field dimensions:", status.Width, "x", status.Height)
	clientLogger.Println("number of players:", len(status.Players))

	for status.Running && status.Players[status.You].Active {

		clientLogger.Println("turn", status.Turn)
		clientLogger.Println("deadline", status.Deadline)

		timingChannel = make(chan time.Time)
		err = getTime(config.timeURL, &serverTime, httpClient)
		if err != nil {
			clientLogger.Fatalln("error receiving time from server")
		}
		go calculateTiming(status.Deadline, serverTime, timingChannel)
		action := config.client.GetAction(*status.Players[status.You], status, timingChannel)
		err = conn.WriteAction(action)
		if err != nil {
			clientLogger.Fatalln("error sending action:", err)
		}

		status, err = conn.ReadStatus()
		if err != nil {
			clientLogger.Fatalln("error reading status:", err)
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
