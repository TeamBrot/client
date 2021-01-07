package main

import (
	"testing"
)

func TestPossibleActions(t *testing.T) {
	type TableEntry struct {
		X             int
		Y             int
		Cells         [][]bool
		OccupiedCells map[Coords]struct{}
		Direction     Direction
		Speed         uint8
		Actions       []Action
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
		{0, 1}: {},
	}

	table := []TableEntry{
		{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 1, Actions: []Action{SpeedUp, TurnRight, ChangeNothing}},
		{X: 0, Y: 1, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 1, Actions: []Action{SpeedUp, TurnRight, ChangeNothing, TurnLeft}},
		{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 2, Actions: []Action{SpeedUp, TurnRight, ChangeNothing, SlowDown}},
		{X: 0, Y: 1, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 2, Actions: []Action{SpeedUp, TurnRight, ChangeNothing, SlowDown}},
		{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Up, Speed: 1, Actions: []Action{TurnRight}},
		{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Left, Speed: 1, Actions: []Action{TurnLeft}},
		{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Up, Speed: 2, Actions: []Action{TurnRight}},
		{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Left, Speed: 2, Actions: []Action{TurnLeft}},
		{X: 4, Y: 4, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 2, Actions: []Action{TurnLeft}},
		{X: 4, Y: 4, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Down, Speed: 2, Actions: []Action{TurnRight}},
		{X: 4, Y: 4, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 2, Actions: []Action{TurnLeft}},
		{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 6, Actions: []Action{}},
		{X: 0, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: nil, Direction: Right, Speed: 1, Actions: []Action{TurnRight}},
		{X: 2, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: nil, Direction: Left, Speed: 1, Actions: []Action{}},
		{X: 0, Y: 1, Turn: 6, Cells: cells2, OccupiedCells: nil, Direction: Right, Speed: 3, Actions: []Action{TurnRight}},
		{X: 0, Y: 2, Turn: 6, Cells: cells2, OccupiedCells: nil, Direction: Right, Speed: 3, Actions: []Action{ChangeNothing, SpeedUp}},
		{X: 0, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: nil, Direction: Right, Speed: 1, Actions: []Action{TurnRight}},
		{X: 0, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: occupiedCells, Direction: Right, Speed: 1, Actions: []Action{ChangeNothing, SpeedUp, TurnRight}},
		{X: 0, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: occupiedCells, Direction: Right, Speed: 3, Actions: []Action{ChangeNothing, SlowDown, SpeedUp, TurnRight}},
		{X: 0, Y: 1, Turn: 1, Cells: cells2, OccupiedCells: occupiedCells, Direction: Right, Speed: 3, Actions: []Action{TurnRight}},
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
		actions := players[1].PossibleActions(status.Cells, entry.Turn, entry.OccupiedCells, true)

		if len(actions) != len(entry.Actions) {
			t.Error("wrong actions, expected", entry.Actions, "got", actions)
			continue
		}
		for _, m1 := range entry.Actions {
			contains := false
			for _, m2 := range actions {
				if m1 == m2 {
					contains = true
				}
			}
			if !contains {
				t.Error("wrong actions, expected", entry.Actions, "got", actions)
			}
		}
	}
}
