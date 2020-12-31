package main

import (
	"testing"
)

func TestParallelMove(t *testing.T) {
	cells := [][]bool{
		{true, true, false, true, true},
		{false, false, false, false, false},
	}
	player := Player{X: 1, Y: 0, Direction: Right, Speed: 1}
	player2 := Player{X: 3, Y: 0, Direction: Left, Speed: 1}
	status := Status{Width: 5, Height: 2, Players: map[uint8]*Player{1: &player, 2: &player2}, Cells: cells, You: 1}
	actions := MinimaxBestActionsTimed(1, 2, &status, nil)
	for _,action := range actions {
		if action == ChangeNothing {
			t.Error("change_nothing should not be a valid action")
		}
	}
	foundTurnRight := false
	for _,action := range actions {
		if action == TurnRight {
			foundTurnRight = true
			break
		}
	}
	if !foundTurnRight {
		t.Error("turn_right should be the best action")
	}
}
