package main

import (
	"net/http"
	"encoding/json"
	"log"
	"time"
)

const timeAPITimeout = 1000 * time.Millisecond

type ServerTime struct {
	Time         time.Time `json:"time"`
	Milliseconds int       `json:"milliseconds"`
}


var httpClient http.Client = http.Client{Timeout: timeAPITimeout}

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
	serverTime, err := getTime(config.timeURL)
	if err != nil {
		log.Println("couldn't reach timing api, try using machine time")
		calculationTime := deadline.Sub(time.Now().UTC())
		calculationTime = time.Duration((calculationTime.Milliseconds() - calculationTimeOffset) * 1000000000)
		if calculationTime > maxCalculationTime*time.Minute {
			return calculationTime, err
		}
		log.Println("the scheduled calculation Time is", calculationTime)
		return calculationTime, nil
	}
	calculationTime := deadline.Sub(serverTime.Time)
	calculationTime = time.Duration((calculationTime.Milliseconds() - int64(serverTime.Milliseconds) - calculationTimeOffset) * 1000000)
	log.Println("the scheduled calculation Time is", calculationTime)
	return calculationTime, nil
}

