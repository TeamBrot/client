package main

import "fmt"

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
	fmt.Println(bestAction)
	return bestAction
}
