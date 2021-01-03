package main

import (
	"log"
	"time"
)

// SmartClient always moves at speed one and chooses right or left if there is an obstacle
type SmartClient struct{}

// GetAction Implementation for SmartClient
func (c SmartClient) GetAction(status *Status, calculationTime time.Duration) Action {
	player := status.Players[status.You]
	action := ChangeNothing
	for _, a := range player.PossibleActions(status.Cells, status.Turn, nil, false) {
		if a == ChangeNothing {
			action = a
			break
		} else if a == TurnLeft || a == TurnRight {
			action = a
		}
	}
	log.Println("using", action)
	return action
}
