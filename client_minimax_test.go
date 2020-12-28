package main

import (
	"testing"
)

func TestPossibleMoves(t *testing.T) {
	type TableEntry struct {
		X             int
		Y             int
		Cells         [][]bool
		OccupiedCells map[Coords]struct{}
		Direction     Direction
		Speed         uint8
		Moves         []Action
		Turn          uint16
	}

	cells := [][]bool{
		{false, false, false, false, false},
		{false, false, false, false, false},
		{false, false, false, false, false},
		{false, false, false, false, false},
		{false, false, false, false, false},
	}

	cells2 := [][]bool{
		{false, true, false, false, false},
		{false, true, true, false, false},
		{false, false, true, false, false},
		{false, false, true, false, false},
		{false, false, false, false, false},
	}

	occupiedCells := map[Coords]struct{}{
		{0,1}: {},
	}

	table := []TableEntry{
		{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 1, Moves: []Action{SpeedUp, TurnRight, ChangeNothing}},
		{X: 0, Y: 1, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 1, Moves: []Action{SpeedUp, TurnRight, ChangeNothing, TurnLeft}},
		{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 2, Moves: []Action{SpeedUp, TurnRight, ChangeNothing, SlowDown}},
		{X: 0, Y: 1, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 2, Moves: []Action{SpeedUp, TurnRight, ChangeNothing, SlowDown}},
		{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Up, Speed: 1, Moves: []Action{TurnRight}},
		{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Left, Speed: 1, Moves: []Action{TurnLeft}},
		{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Up, Speed: 2, Moves: []Action{TurnRight}},
		{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Left, Speed: 2, Moves: []Action{TurnLeft}},
		{X: 4, Y: 4, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 2, Moves: []Action{TurnLeft}},
		{X: 4, Y: 4, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Down, Speed: 2, Moves: []Action{TurnRight}},
		{X: 4, Y: 4, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 2, Moves: []Action{TurnLeft}},
		{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 6, Moves: []Action{}},
		{X: 0, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: nil, Direction: Right, Speed: 1, Moves: []Action{TurnRight}},
		{X: 2, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: nil, Direction: Left, Speed: 1, Moves: []Action{}},
		{X: 0, Y: 1, Turn: 6, Cells: cells2, OccupiedCells: nil, Direction: Right, Speed: 3, Moves: []Action{TurnRight}},
		{X: 0, Y: 2, Turn: 6, Cells: cells2, OccupiedCells: nil, Direction: Right, Speed: 3, Moves: []Action{ChangeNothing, SpeedUp}},
		{X: 0, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: nil, Direction: Right, Speed: 1, Moves: []Action{TurnRight}},
		{X: 0, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: occupiedCells, Direction: Right, Speed: 1, Moves: []Action{ChangeNothing, SpeedUp, TurnRight}},
		{X: 0, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: occupiedCells, Direction: Right, Speed: 3, Moves: []Action{ChangeNothing, SlowDown, SpeedUp, TurnRight}},
		{X: 0, Y: 1, Turn: 1, Cells: cells2, OccupiedCells: occupiedCells, Direction: Right, Speed: 3, Moves: []Action{TurnRight}},
	}

	players := map[int]*Player{
		1: {},
	}
	status := Status{Width: 5, Height: 5, Turn: 1}

	for _, entry := range table {
		players[1].X = uint16(entry.X)
		players[1].Y = uint16(entry.Y)
		players[1].Direction = entry.Direction
		players[1].Speed = uint8(entry.Speed)
		status.Cells = entry.Cells
		status.Turn = uint16(entry.Turn)
		moves := players[1].PossibleMoves(status.Cells, entry.Turn, entry.OccupiedCells, true)

		if len(moves) != len(entry.Moves) {
			t.Error("wrong moves, expected", entry.Moves, "got", moves)
			continue
		}
		for _, m1 := range entry.Moves {
			contains := false
			for _, m2 := range moves {
				if m1 == m2 {
					contains = true
				}
			}
			if !contains {
				t.Error("wrong moves, expected", entry.Moves, "got", moves)
			}
		}
	}
}

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
