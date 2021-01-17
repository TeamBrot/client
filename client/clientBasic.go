package main

import (
	"time"
)

// BasicClient always moves at speed one and chooses right or left if there is an obstacle
type BasicClient struct{}

// GetAction Implementation for BasicClient
func (c BasicClient) GetAction(status *Status, calculationTime time.Duration) Action {
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
	return action
}
