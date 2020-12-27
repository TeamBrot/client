package main

//func TestNewOccupiedCells(t *testing.T) {
//status := Status{Width: 5, Height: 10}
//occupiedCells := newOccupiedCells(&status)
//if len(occupiedCells) != status.Height {
//t.Error("height of occupiedCells does not match, got", len(occupiedCells), "expected", status.Height)
//}
//for _, m := range occupiedCells {
//if len(m) != status.Width {
//t.Error("width of occupiedCells does not match, got", len(m), "expected", status.Width)
//}
//for _, n := range m {
//if n != false {
//t.Error("true value in occupiedCells after initialization")
//}
//}
//}
//}

//func TestPossibleMoves(t *testing.T) {
//type TableEntry struct {
//X             int
//Y             int
//Cells         [][]int
//OccupiedCells [][]bool
//Direction     Direction
//Speed         int
//Moves         []Action
//Turn          int
//}

//cells := [][]int{
//{0, 0, 0, 0, 0},
//{0, 0, 0, 0, 0},
//{0, 0, 0, 0, 0},
//{0, 0, 0, 0, 0},
//{0, 0, 0, 0, 0},
//}

//cells2 := [][]int{
//{0, 2, 0, 0, 0},
//{0, 2, 2, 0, 0},
//{0, 0, 2, 0, 0},
//{0, 0, 2, 0, 0},
//{0, 0, 0, 0, 0},
//}

//occupiedCells := [][]bool{
//{false, true, false, false, false},
//{false, false, false, false, false},
//{false, false, false, false, false},
//{false, false, false, false, false},
//{false, false, false, false, false},
//}

//table := []TableEntry{
//{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 1, Moves: []Action{SpeedUp, TurnRight, ChangeNothing}},
//{X: 0, Y: 1, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 1, Moves: []Action{SpeedUp, TurnRight, ChangeNothing, TurnLeft}},
//{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 2, Moves: []Action{SpeedUp, TurnRight, ChangeNothing, SlowDown}},
//{X: 0, Y: 1, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 2, Moves: []Action{SpeedUp, TurnRight, ChangeNothing, SlowDown}},
//{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Up, Speed: 1, Moves: []Action{TurnRight}},
//{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Left, Speed: 1, Moves: []Action{TurnLeft}},
//{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Up, Speed: 2, Moves: []Action{TurnRight}},
//{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Left, Speed: 2, Moves: []Action{TurnLeft}},
//{X: 4, Y: 4, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 2, Moves: []Action{TurnLeft}},
//{X: 4, Y: 4, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Down, Speed: 2, Moves: []Action{TurnRight}},
//{X: 4, Y: 4, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 2, Moves: []Action{TurnLeft}},
//{X: 0, Y: 0, Turn: 1, Cells: cells, OccupiedCells: nil, Direction: Right, Speed: 6, Moves: []Action{}},
//{X: 0, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: nil, Direction: Right, Speed: 1, Moves: []Action{TurnRight}},
//{X: 2, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: nil, Direction: Left, Speed: 1, Moves: []Action{}},
//{X: 0, Y: 1, Turn: 6, Cells: cells2, OccupiedCells: nil, Direction: Right, Speed: 3, Moves: []Action{TurnRight}},
//{X: 0, Y: 2, Turn: 6, Cells: cells2, OccupiedCells: nil, Direction: Right, Speed: 3, Moves: []Action{ChangeNothing, SpeedUp}},
//{X: 0, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: nil, Direction: Right, Speed: 1, Moves: []Action{TurnRight}},
//{X: 0, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: occupiedCells, Direction: Right, Speed: 1, Moves: []Action{ChangeNothing, SpeedUp, TurnRight}},
//{X: 0, Y: 0, Turn: 1, Cells: cells2, OccupiedCells: occupiedCells, Direction: Right, Speed: 3, Moves: []Action{ChangeNothing, SlowDown, SpeedUp, TurnRight}},
//{X: 0, Y: 1, Turn: 1, Cells: cells2, OccupiedCells: occupiedCells, Direction: Right, Speed: 3, Moves: []Action{TurnRight}},
//}
//players := map[int]*Player{
//1: {Active: true},
//}
//status := Status{Width: 5, Height: 5, Running: true, Turn: 1, Players: players}

//for _, entry := range table {
//players[1].X = entry.X
//players[1].Y = entry.Y
//players[1].Direction = entry.Direction
//players[1].Speed = entry.Speed
//status.Cells = entry.Cells
//status.Turn = entry.Turn
//moves := Moves(&status, players[1], entry.OccupiedCells)

//if len(moves) != len(entry.Moves) {
//t.Error("wrong moves, expected", entry.Moves, "got", moves)
//continue
//}
//for _, m1 := range entry.Moves {
//contains := false
//for _, m2 := range moves {
//if m1 == m2 {
//contains = true
//}
//}
//if !contains {
//t.Error("wrong moves, expected", entry.Moves, "got", moves)
//}
//}
//}
//}

// func TestDoMove(t *testing.T) {

// 	type TableEntry struct {
// 		X1                int
// 		Y1                int
// 		X2                int
// 		Y2                int
// 		Cells             [][]int
// 		CellsCopy         [][]int
// 		OccupiedCells     [][]bool
// 		OccupiedCellsCopy [][]bool
// 		Survive           bool
// 		Direction         Direction
// 		Speed             int
// 		Speed2            int
// 		Move              Action
// 		Turn              int
// 	}

// 	cells := [][]int{
// 		{0, 0, 0, 0, 0},
// 		{0, 0, 0, 0, 0},
// 		{0, 0, 0, 0, 0},
// 		{0, 0, 0, 0, 0},
// 		{0, 0, 0, 0, 0},
// 	}

// 	cellsCopy := [][]int{
// 		{0, 0, 0, 0, 0},
// 		{0, 0, 0, 0, 0},
// 		{0, 0, 0, 0, 0},
// 		{0, 0, 0, 0, 0},
// 		{0, 0, 0, 0, 0},
// 	}

// 	cells2 := [][]int{
// 		{0, 2, 0, 0, 0},
// 		{0, 2, 2, 0, 0},
// 		{0, 0, 2, 0, 0},
// 		{0, 0, 2, 0, 0},
// 		{0, 0, 0, 0, 0},
// 	}

// 	cells2Copy := [][]int{
// 		{0, 2, 0, 0, 0},
// 		{0, 2, 2, 0, 0},
// 		{0, 0, 2, 0, 0},
// 		{0, 0, 2, 0, 0},
// 		{0, 0, 0, 0, 0},
// 	}

// 	occupiedCells := [][]bool{
// 		{false, true,  false, false, false},
// 		{false, true,  false, false, false},
// 		{false, false, false, false, false},
// 		{false, false, false, false, false},
// 		{false, false, false, false, false},
// 	}

// 	occupiedCellsCopy := [][]bool{
// 		{false, true,  false, false, false},
// 		{false, true,  false, false, false},
// 		{false, false, false, false, false},
// 		{false, false, false, false, false},
// 		{false, false, false, false, false},
// 	}

// 	table := []TableEntry{
// 		{X1: 0, Y1: 0, X2: 0, Y2: 1, Cells: cells, CellsCopy: cellsCopy, OccupiedCells: occupiedCells, OccupiedCellsCopy: occupiedCellsCopy, Survive: true, Direction: Down, Speed: 1, Speed2: 1, Move: ChangeNothing, Turn: 1},
// 		{X1: 0, Y1: 0, X2: 0, Y2: 4, Cells: cells, CellsCopy: cellsCopy, OccupiedCells: occupiedCells, OccupiedCellsCopy: occupiedCellsCopy, Survive: true, Direction: Down, Speed: 4, Speed2: 4, Move: ChangeNothing, Turn: 1},
// 		{X1: 0, Y1: 2, X2: 3, Y2: 2, Cells: cells2, CellsCopy: cells2Copy, OccupiedCells: occupiedCells, OccupiedCellsCopy: occupiedCellsCopy, Survive: true, Direction: Right, Speed: 3, Speed2: 3, Move: ChangeNothing, Turn: 6},
// 		{X1: 0, Y1: 0, X2: 1, Y2: 0, Cells: cells2, CellsCopy: cells2Copy, OccupiedCells: occupiedCells, OccupiedCellsCopy: occupiedCellsCopy, Survive: false, Direction: Right, Speed: 1, Speed2: 1, Move: ChangeNothing, Turn: 1},
// 		{X1: 4, Y1: 1, X2: 1, Y2: 0, Cells: cells2, CellsCopy: cells2Copy, OccupiedCells: occupiedCells, OccupiedCellsCopy: occupiedCellsCopy, Survive: false, Direction: Left, Speed: 3, Speed2: 3, Move: ChangeNothing, Turn: 6},
// 		{X1: 4, Y1: 1, X2: 1, Y2: 0, Cells: cells2, CellsCopy: cells2Copy, OccupiedCells: occupiedCells, OccupiedCellsCopy: occupiedCellsCopy, Survive: false, Direction: Left, Speed: 2, Speed2: 3, Move: SpeedUp, Turn: 6},
// 		{X1: 4, Y1: 1, X2: 0, Y2: 1, Cells: cells2, CellsCopy: cells2Copy, OccupiedCells: occupiedCells, OccupiedCellsCopy: occupiedCellsCopy, Survive: true, Direction: Left, Speed: 3, Speed2: 4, Move: SpeedUp, Turn: 6},
// 		{X1: 1, Y1: 4, X2: 1, Y2: 0, Cells: cells2, CellsCopy: cells2Copy, OccupiedCells: occupiedCells, OccupiedCellsCopy: occupiedCellsCopy, Survive: false, Direction: Up, Speed: 3, Speed2: 3, Move: ChangeNothing, Turn: 6},
// 	}

// 	players := map[int]*Player{
// 		1: {Active: true},
// 	}
// 	status := Status{Width: 5, Height: 5, Running: true, Turn: 1, Players: players}

// 	for _, entry := range table {
// 		players[1].X = entry.X1
// 		players[1].Y = entry.Y1
// 		players[1].Direction = entry.Direction
// 		players[1].Speed = entry.Speed
// 		status.Cells = entry.Cells
// 		status.Turn = entry.Turn
// 		survive := doMove(&status, 1, entry.Move, entry.OccupiedCells)
// 		if survive != entry.Survive {
// 			t.Error("real and expected survival does not match, expected", entry.Survive, "got", survive)
// 		}
// 		if survive {
// 			if status.Players[1].X != entry.X2 || status.Players[1].Y != entry.Y2 {
// 				t.Error("player is not at specified position while surviving, got x =", status.Players[1].X, "and y =", status.Players[1].Y, " but expected x =", entry.X2, "and y =", entry.Y2)
// 			}
// 			if status.Players[1].Speed != entry.Speed2 {
// 				t.Error("player is not going in the specified speed, got", status.Players[1].Speed, "but expected", entry.Speed2)
// 			}
// 			undoMove(&status, status.Players[1], entry.Move, entry.OccupiedCells)
// 			if status.Players[1].X != entry.X1 || status.Players[1].Y != entry.Y1 {
// 				t.Error("player is not at specified position, got x =", status.Players[1].X, "and y =", status.Players[1].Y, " but expected x =", entry.X1, "and y =", entry.Y1)
// 			}
// 			if status.Players[1].Direction != entry.Direction {
// 				t.Error("player is not going in the specified direction, got", status.Players[1].Direction, "but expected", entry.Direction)
// 			}
// 			if status.Players[1].Speed != entry.Speed {
// 				t.Error("player is not going in the specified speed, got", status.Players[1].Speed, "but expected", entry.Speed)
// 			}
// 		} else {
// 			if status.Players[1].X != entry.X1 || status.Players[1].Y != entry.Y1 {
// 				t.Error("player is not at specified position while dying, got x =", status.Players[1].X, "and y =", status.Players[1].Y, " but expected x =", entry.X1, "and y =", entry.Y1)
// 			}
// 		}
// 		for i := range status.Cells {
// 			for j := range status.Cells[i] {
// 				if status.Cells[i][j] != entry.CellsCopy[i][j] {
// 					t.Error("cells is not the same as before, got", status.Cells[i][j], "expected", entry.CellsCopy[i][j], "at x =", j, "and y =", i)
// 				}
// 			}
// 		}
// 		for i := range entry.OccupiedCells {
// 			for j := range entry.OccupiedCells[i] {
// 				if entry.OccupiedCells[i][j] != entry.OccupiedCellsCopy[i][j] {
// 					t.Error("occupiedCells is not the same as before, got", entry.OccupiedCells[i][j], "expected", entry.OccupiedCellsCopy[i][j], "at x =", j, "and y =", i)
// 				}
// 			}
// 		}
// 	}
// }
