package main

import (
	"log"
	"time"
)

// SmartClient always moves at speed one and chooses right or left if there is an obstacle
type SmartClient struct{}

// GetAction Implementation for SmartClient
func (c SmartClient) GetAction(player Player, status *Status, calculationTime time.Duration) Action {
	var bestAction Action
	board := status.Cells
	switch player.Direction {
	case Up:
		if player.Y == 0 {
			if player.X+1 >= len(board[0]) {
				bestAction = TurnLeft
			} else if player.X == 0 {
				bestAction = TurnRight
			} else if board[player.Y][player.X+1] == 0 {
				bestAction = TurnRight
			} else {
				bestAction = TurnLeft
			}
		} else if board[player.Y-1][player.X] != 0 {
			if player.X+1 >= len(board[0]) {
				bestAction = TurnLeft
			} else if player.X == 0 {
				bestAction = TurnRight
			} else if board[player.Y][player.X+1] == 0 {
				bestAction = TurnRight
			} else {
				bestAction = TurnLeft
			}
		} else {
			bestAction = ChangeNothing
		}
	case Down:
		if player.Y+1 >= len(board) {
			if player.X == 0 {
				bestAction = TurnLeft
			} else if player.X+1 >= len(board[0]) {
				bestAction = TurnRight
			} else if board[player.Y][player.X+1] == 0 {
				bestAction = TurnLeft
			} else {
				bestAction = TurnRight
			}
		} else if board[player.Y+1][player.X] != 0 {
			if player.X == 0 {
				bestAction = TurnLeft
			} else if player.X+1 >= len(board[0]) {
				bestAction = TurnRight
			} else if board[player.Y][player.X+1] == 0 {
				bestAction = TurnLeft
			} else {
				bestAction = TurnRight
			}
		} else {
			bestAction = ChangeNothing
		}
	case Left:
		if player.X == 0 {
			if player.Y == 0 {
				bestAction = TurnLeft
			} else if player.Y+1 >= len(board) {
				bestAction = TurnRight
			} else if board[player.Y+1][player.X] == 0 {
				bestAction = TurnLeft
			} else {
				bestAction = TurnRight
			}
		} else if board[player.Y][player.X-1] != 0 {
			if player.Y == 0 {
				bestAction = TurnLeft
			} else if player.Y+1 >= len(board) {
				bestAction = TurnRight
			} else if board[player.Y+1][player.X] == 0 {
				bestAction = TurnLeft
			} else {
				bestAction = TurnRight
			}
		} else {
			bestAction = ChangeNothing
		}
	case Right:
		if player.X+1 >= len(board[0]) {
			if player.Y == 0 {
				bestAction = TurnRight
			} else if player.Y+1 >= len(board) {
				bestAction = TurnLeft
			} else if board[player.Y+1][player.X] == 0 {
				bestAction = TurnRight
			} else {
				bestAction = TurnLeft
			}
		} else if board[player.Y][player.X+1] != 0 {
			if player.Y == 0 {
				bestAction = TurnRight
			} else if player.Y+1 >= len(board) {
				bestAction = TurnLeft
			} else if board[player.Y+1][player.X] == 0 {
				bestAction = TurnRight
			} else {
				bestAction = TurnLeft
			}
		} else {
			bestAction = ChangeNothing
		}
	}
	log.Println("using", bestAction)
	return bestAction
}
