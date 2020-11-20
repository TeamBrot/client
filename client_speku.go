package main

import "fmt"

//Field to store probabilities
type Field struct {
	Cells   [][]float64
	Players map[int]*Player
}

func combineFields(field1 Field, field2 Field) {

}

func convertCellsToField(status *Status) *Field {
	cells := status.Cells
	height := status.Height
	width := status.Width

	fieldCells := make([][]float64, height+2)
	for i := range fieldCells {
		fieldCells[i] = make([]float64, width+2)
	}
	for i := range cells {
		for j := range cells[i] {
			if cells[i][j] != 0 {
				fieldCells[i+1][j+1] = -1.0
			} else {
				fieldCells[i+1][j+1] = 0.0
			}
		}
	}
	for i := 0; i < height+2; i++ {
		for j := 0; j < width+2; j++ {
			if i == 0 || i == height+1 || j == 0 || j == width+1 {
				fieldCells[i][j] = -1.0
			}
		}
	}
	playerMap := make(map[int]*Player)
	for i, player := range status.Players {
		player.Y = player.Y + 1
		player.X = player.X + 1
		playerMap[i] = player
	}
	field := Field{Cells: fieldCells, Players: playerMap}
	return &field
}

func simulateMove(field *Field, playerID int, probability float64, action Action, turn int) *Field {
	player := copyPlayer(field.Players[playerID])
	if action == SpeedUp {
		if player.Speed != 10 {
			player.Speed++
		} else {
			return field
		}
	} else if action == SlowDown {
		if player.Speed != 1 {
			player.Speed--
		} else {
			return field
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
			if field.Cells[player.Y][player.X] >= 0 {
				field.Cells[player.Y][player.X] = field.Cells[player.Y][player.X] + probability
			} else {
				for j := i; i > 1; i-- {
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
					if action == SpeedUp {
						player.Speed--
					} else if action == SlowDown {
						player.Speed++
					} else if action == TurnLeft {
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
					} else if action == TurnRight {
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
					}
				}
				return field
			}

		}
	}
	field.Players[len(field.Players)+1] = player
	return field
}

// SpekuClient is a client implementation that uses Minimax to decide what to do next
type SpekuClient struct{}

// GetAction implements the Client interface
//TODO: use player information
func (c SpekuClient) GetAction(player Player, status *Status) Action {
	turns := status.Turn
	field := convertCellsToField(status)
	for c := range field.Players {
		for _, action := range Actions {
			field = simulateMove(field, c, 0.2, action, turns)
		}
		turns++
	}
	fmt.Println(*field)
	return ChangeNothing
}
