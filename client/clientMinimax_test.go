package main

import (
	"testing"
	"reflect"
)

func TestParallelMove(t *testing.T) {
	cells := [][]bool{
		{true, true, false, true, true},
		{false, false, false, false, false},
	}
	player := Player{X: 1, Y: 0, Direction: Right, Speed: 1}
	player2 := Player{X: 3, Y: 0, Direction: Left, Speed: 1}
	status := Status{Width: 5, Height: 2, Players: map[uint8]*Player{1: &player, 2: &player2}, Cells: cells, You: 1}
	actions := MinimaxBestActions(1, []uint8{2}, &status, nil)
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

func TestCombineScoreMapIdentity(t *testing.T) {
	scoreMap := map[Action]int{ChangeNothing: 0, TurnRight: 1}
	resultScoreMap := combineScoreMaps([]map[Action]int{scoreMap, scoreMap})
	if !reflect.DeepEqual(resultScoreMap, scoreMap) {
		t.Errorf("score maps not equal")
	}
}

func TestCombineScoreMapMinimum(t *testing.T) {
	scoreMap1 := map[Action]int{ChangeNothing: 2, TurnRight: 1}
	scoreMap2 := map[Action]int{ChangeNothing: 0, TurnRight: 2}
	resultScoreMap := map[Action]int{ChangeNothing: 0, TurnRight: 1}
	if !reflect.DeepEqual(resultScoreMap, combineScoreMaps([]map[Action]int{scoreMap1, scoreMap2})) {
		t.Errorf("score maps not equal")
	}
}

func TestBestActionsFromScoreMap(t *testing.T) {
	scoreMap1 := map[Action]int{ChangeNothing: 2, TurnRight: 1}
	bestActions1 := []Action{ChangeNothing}
	if !reflect.DeepEqual(bestActions1, bestActionsFromScoreMap(scoreMap1)) {
		t.Error("best actions not equal", bestActions1, bestActionsFromScoreMap(scoreMap1))
	}

	scoreMap2 := map[Action]int{ChangeNothing: 2, TurnRight: 2, TurnLeft: 2, SpeedUp: -1}
	bestActions2 := []Action{ChangeNothing, TurnLeft, TurnRight}
	if !reflect.DeepEqual(bestActionsFromScoreMap(scoreMap2), bestActions2) {
		t.Error("best actions are not equal", bestActions2, bestActionsFromScoreMap(scoreMap2))
	}
}

