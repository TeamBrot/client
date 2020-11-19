package main

import "fmt"

//Field to store probabilities
type Field struct {
	Cells   [][]float64
	Players map[int]*Player
}

func combineFields(field1 Field, field2 Field) {

}

func convertCellsToField(cells [][]int, width int, height int) *Field {
	fieldCells := make([][]float64, height)
	for i := range fieldCells {
		fieldCells[i] = make([]float64, width)
	}
	for i := range cells {
		for j := range cells[i] {
			if cells[i][j] != 0 {
				fieldCells[i][j] = -1.0
			} else {
				fieldCells[i][j] = 0.0
			}
		}
	}
	field := Field{Cells: fieldCells, Players: make(map[int]*Player)}
	return &field
}

func simulateMove(field *Field, playerID int, probability float64, action Action, turn int) *Field {
	player := field.Players[playerID]
	if action == SpeedUp {
		if player.Speed != 10 {
			player.Speed++
		}
	} else if action == SlowDown {
		if player.Speed != 1 {
			player.Speed--
		}
	} else if action == TurnLeft {
		switch player.Direction {
		case Left:
			player.Direction = Down
			break
		case Down:
			player.Direction = Right
			break
		case Right:
			player.Direction = Up
			break
		case Up:
			player.Direction = Left
			break
		}
	} else if action == TurnRight {
		switch player.Direction {
		case Left:
			player.Direction = Up
			break
		case Down:
			player.Direction = Left
			break
		case Right:
			player.Direction = Down
			break
		case Up:
			player.Direction = Right
			break
		}
	}

	jump := turn%6 == 0
	for i := 1; i <= player.Speed; i++ {
		if player.Direction == Up {
			player.Y--
		} else if player.Direction == Down {
			player.Y++
		} else if player.Direction == Right {
			player.X++
		} else if player.Direction == Left {
			player.X--
		}

		if !jump || i == 1 || i == player.Speed {
			if field.Cells[player.Y][player.X] > 0 {
				field.Cells[player.Y][player.X] = field.Cells[player.Y][player.X] + probability
			} else {
				for j := i; i > 0; i-- {
					if player.Direction == Up {
						player.Y++
					} else if player.Direction == Down {
						player.Y--
					} else if player.Direction == Right {
						player.X--
					} else if player.Direction == Left {
						player.X++
					}
					if !jump || j == 1 || j == player.Speed {
						field.Cells[player.Y][player.X] = field.Cells[player.Y][player.X] - probability
					}
				}
				break
			}

		}
	}
	return field
}

// SpekuClient is a client implementation that uses Minimax to decide what to do next
type SpekuClient struct{}

// GetAction implements the Client interface
//TODO: use player information
func (c SpekuClient) GetAction(player Player, status *Status) Action {
	field := convertCellsToField(status.Cells, status.Width, status.Height)
	field.Players[2] = status.Players[2]
	fmt.Println(*simulateMove(field, 2, 0.2, ChangeNothing, status.Turn))

	return ChangeNothing
}
