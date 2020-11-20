package main

import (
	"fmt"
)

//Field to store probabilities
type Field struct {
	Cells   [][]float64
	Players map[int]*SimPlayer
}

type Coords struct {
	Y, X int
}

//SimPlayer to add a new array of visited cells
type SimPlayer struct {
	X            int `json:"x"`
	Y            int `json:"y"`
	Direction    Direction
	Speed        int `json:"speed"`
	visitedCells map[Coords]struct{}
}

func combineFields(field1 Field, field2 Field) {

}

func convertCellsToField(status *Status) *Field {
	cells := status.Cells
	height := status.Height
	width := status.Width

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
	simPlayerMap := make(map[int]*SimPlayer)
	count := 1
	for i, player := range status.Players {
		if i != status.You {
			newSimPlayer := SimPlayer{X: player.X, Y: player.Y, Direction: player.Direction, Speed: player.Speed, visitedCells: make(map[Coords]struct{})}
			simPlayerMap[count] = &newSimPlayer
			count++
		}
	}
	field := Field{Cells: fieldCells, Players: simPlayerMap}
	return &field
}

func simulateMove(field *Field, playerID int, probability float64, action Action, turn int, limit int) (*Field, int) {
	player := copyPlayer(field.Players[playerID])
	if action == SpeedUp {
		if player.Speed != 10 {
			player.Speed++
		} else {
			return field, 0
		}
	} else if action == SlowDown {
		if player.Speed != 1 {
			player.Speed--
		} else {
			return field, 0
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

			inField := player.Y >= 0 && player.Y < len(field.Cells) && player.X >= 0 && player.X < len(field.Cells[0])
			coordsNow := Coords{Y: player.Y, X: player.X}
			_, fieldVisited := player.visitedCells[coordsNow]

			if inField && field.Cells[player.Y][player.X] >= 0 && !fieldVisited {
				field.Cells[player.Y][player.X] = field.Cells[player.Y][player.X] + probability
				player.visitedCells[coordsNow] = struct{}{}
			} else {
				for j := i; j > 1; j-- {
					if player.Direction == Up {
						player.Y++
					} else if player.Direction == Down {
						player.Y--
					} else if player.Direction == Right {
						player.X--
					} else if player.Direction == Left {
						player.X++
					}
					coordsNow := Coords{Y: player.Y, X: player.X}
					if (!jump || j == 1 || j == player.Speed) && inField {
						field.Cells[player.Y][player.X] = field.Cells[player.Y][player.X] - probability
						if !fieldVisited {
							delete(player.visitedCells, coordsNow)
						}
						fieldVisited = false
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
				return field, 0
			}

		}
	}
	if len(field.Players)+1 > limit+10 {
		field.Players[len(field.Players)+1] = nil
	} else {
		field.Players[len(field.Players)+1] = player
	}
	return field, 1
}
func copyPlayer(player *SimPlayer) *SimPlayer {
	var p SimPlayer
	p.Direction = player.Direction
	p.Speed = player.Speed
	p.X = player.X
	p.Y = player.Y
	p.visitedCells = make(map[Coords]struct{})
	for k, v := range player.visitedCells {
		p.visitedCells[k] = v
	}
	return &p
}

// SpekuClient is a client implementation that uses Minimax to decide what to do next
type SpekuClient struct{}

// GetAction implements the Client interface
//TODO: use player information
func (c SpekuClient) GetAction(player Player, status *Status) Action {
	turns := 0
	field := convertCellsToField(status)
	limit := 2500000
	lenTurn := 1
	move := 0
	var movemade int
	for i := 1; i <= len(field.Players); i++ {
		if i > lenTurn {
			turns++
			lenTurn = lenTurn + move
			move = 0
		}
		probability := 1.0 / (5.0 * float64(i))
		for _, action := range Actions {
			field, movemade = simulateMove(field, i, probability, action, turns, limit)
			move = move + movemade
		}
		field.Players[i] = nil
		if i > limit {
			fmt.Println("Ich habe abgebrochen")
			break
		}
	}
	fmt.Println(field.Cells)
	fmt.Println(turns)
	return ChangeNothing
}
