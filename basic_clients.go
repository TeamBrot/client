package main

import "log"

func leftRight(me Player, board [][]int, goDirection Action) Action {
	var action Action
	switch me.Direction {
	case Up:
		if me.Y == 0 {
			action = goDirection
			break
		} else if board[me.Y-1][me.X] != 0 {
			action = goDirection
		} else {
			action = ChangeNothing
		}
	case Down:
		if me.Y+1 >= len(board) {
			action = goDirection
			break
		} else if board[me.Y+1][me.X] != 0 {
			action = goDirection
		} else {
			action = ChangeNothing
		}
	case Right:
		if me.X+1 >= len(board[0]) {
			action = goDirection
			break
		} else if board[me.Y][me.X+1] != 0 {
			action = goDirection
		} else {
			action = ChangeNothing
		}
	case Left:
		if me.X == 0 {
			action = goDirection
			break
		} else if board[me.Y][me.X-1] != 0 {
			action = goDirection
		} else {
			action = ChangeNothing
		}
	}
	log.Println(action)
	return action
}

// LeftClient always goes left when there is an obstacle
type LeftClient struct{}

// GetAction implementation for LeftClient
func (c LeftClient) GetAction(player Player, status *Status) Action {
	var bestAction Action
	bestAction = leftRight(player, status.Cells, TurnLeft)
	return bestAction
}

// RightClient always goes right when there is an obstacle
type RightClient struct{}

// GetAction implementation for RightClient
func (c RightClient) GetAction(player Player, status *Status) Action {
	var bestAction Action
	bestAction = leftRight(player, status.Cells, TurnRight)
	return bestAction
}

// SmartClient always moves at speed one and chooses right or left if there is an obstacle
type SmartClient struct{}

// GetAction Implementation for SmartClient
func (c SmartClient) GetAction(player Player, status *Status) Action {
	var bestAction Action
	board := status.Cells
	log.Println(player.Direction)
	switch player.Direction {
	case "up":
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
	case "down":
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
	case "left":
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
	case "right":
		if player.X+1 >= len(board) {
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
	log.Println(bestAction)
	return bestAction
}
