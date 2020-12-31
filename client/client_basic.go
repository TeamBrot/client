package main

import (
	"log"
	"time"
)

// SmartClient always moves at speed one and chooses right or left if there is an obstacle
type SmartClient struct{}

// GetAction Implementation for SmartClient
func (c SmartClient) GetAction(player Player, status *Status, calculationTime time.Duration) Action {
	action := ChangeNothing
	for _, a := range player.PossibleMoves(status.Cells, status.Turn, nil, false) {
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
