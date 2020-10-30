package main

import "log"

func left_right(me Player, board [][]int, go_direction Action) Action {
	var action Action
	change_nothing := ChangeNothing
	switch me.Direction {
	case "up":
		if me.Y == 0 {
			action = go_direction
			break
		} else if board[me.Y-1][me.X] != 0 {
			action = go_direction
		} else {
			action = change_nothing
		}
	case "down":
		if me.Y+1 >= len(board) {
			action = go_direction
			break
		} else if board[me.Y+1][me.X] != 0 {
			action = go_direction
		} else {
			action = change_nothing
		}
	case "right":
		if me.X+1 >= len(board[0]) {
			action = go_direction
			break
		} else if board[me.Y][me.X+1] != 0 {
			action = go_direction
		} else {
			action = change_nothing
		}
	case "left":
		if me.X == 0 {
			action = go_direction
			break
		} else if board[me.Y][me.X-1] != 0 {
			action = go_direction
		} else {
			action = change_nothing
		}
	}
	log.Println(action)
	return action
}

type LeftClient struct{}

func (c LeftClient) GetAction(player Player, status *Status) Action {
	var bestAction Action
	bestAction = left_right(player, status.Cells, TurnLeft)
	return bestAction
}

type RightClient struct{}

func (c RightClient) GetAction(player Player, status *Status) Action {
	var bestAction Action
	bestAction = left_right(player, status.Cells, TurnRight)
	return bestAction
}

type SmartClient struct{}

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
