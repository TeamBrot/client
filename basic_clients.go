package main

import "fmt"

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
	fmt.Println(bestAction)
	return bestAction
}
