package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

const timeAPIRequestTimeout = 1000 * time.Millisecond

// If getting the time from the timing api fails and the calculation time calculated is more than this value in minutes away the program will throw an error
const maxCalculationTime = 2 * time.Minute

// This value is specified in milliseconds and is a the expected time which actions take to be sent to the server
const calculationTimeOffset = 200

type ServerTime struct {
	Time         time.Time `json:"time"`
	Milliseconds int       `json:"milliseconds"`
}

var httpClient http.Client = http.Client{Timeout: timeAPIRequestTimeout}

//gets the current server Time via the specified API
func getTime(url string) (ServerTime, error) {
	var timeFromServer ServerTime
	r, err := httpClient.Get(url)
	if err != nil {
		return timeFromServer, err
	}
	defer r.Body.Close()
	json.NewDecoder(r.Body).Decode(&timeFromServer)
	if timeFromServer.Time.IsZero() {
		return timeFromServer, errors.New("invalid time from api")
	}
	return timeFromServer, nil
}

//Sends Signals to the Client after a specified amount of time has passed
func computeCalculationTime(deadline time.Time, config Config, errorLogger *log.Logger) (time.Duration, error) {
	serverTime, err := getTime(config.TimeURL)
	if err != nil {
		errorLogger.Println("couldn't reach timing api, try using machine time")
		calculationTime := deadline.Sub(time.Now().UTC())
		calculationTime = time.Duration((calculationTime.Milliseconds() - calculationTimeOffset) * 1000000000)
		if calculationTime > maxCalculationTime {
			return calculationTime, errors.New("couldn't reach timing api and deadline is more than the specified maxCalculationTime away")
		}
		log.Println("the scheduled calculation Time is", calculationTime)
		log.Println("This is only a assumption, based on the local machine tim!")
		return calculationTime, nil
	}
	calculationTime := deadline.Sub(serverTime.Time)
	calculationTime = time.Duration((calculationTime.Milliseconds() - int64(serverTime.Milliseconds) - calculationTimeOffset) * 1000000)
	log.Println("the scheduled calculation Time is", calculationTime)
	return calculationTime, nil
}
