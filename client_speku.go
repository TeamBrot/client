package main

import (
	"fmt"
	"math"
	"time"
)

//Field to store probabilities
type Field struct {
	Cells   [][]float64
	Players []*SimPlayer
}

//Coords store the coordinates of a player
type Coords struct {
	Y, X int
}

//SimPlayer to add a new array of visited cells
type SimPlayer struct {
	X            int `json:"x"`
	Y            int `json:"y"`
	Direction    Direction
	Speed        int `json:"speed"`
	probability  float64
	visitedCells map[Coords]struct{}
}

func combineFields(field1 Field, field2 Field) {

}

func convertCellsToField(status *Status) []*Field {
	cells := status.Cells
	height := status.Height
	width := status.Width
	fieldArray := make([]*Field, 0)
	for i, player := range status.Players {

		if i != status.You {
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
			simPlayerMap := make([]*SimPlayer, 0)
			simPlayerMap = append(simPlayerMap, nil)
			newSimPlayer := SimPlayer{X: player.X, Y: player.Y, Direction: player.Direction, Speed: player.Speed, visitedCells: make(map[Coords]struct{})}
			simPlayerMap = append(simPlayerMap, &newSimPlayer)

			field := Field{Cells: fieldCells, Players: simPlayerMap}
			fieldArray = append(fieldArray, &field)
		}
	}
	return fieldArray
}

func simulateMove(field *Field, playerID int, probability float64, action Action, turn int, limit int) (*Field, int) {
	if field.Players[playerID] == nil {
		return field, 0
	}
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
		field.Players = append(field.Players, nil)
	} else {
		field.Players = append(field.Players, player)
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
	for k := range player.visitedCells {
		p.visitedCells[k] = struct{}{}
	}
	return &p
}

func playerEqual(player1 *SimPlayer, player2 *SimPlayer) bool {
	if player1.Y != player2.Y || player1.X != player2.X {
		return false
	} else if player1.Speed != player2.Speed {
		return false
	} else if player1.Direction != player2.Direction {
		return false
	} else if len(player1.visitedCells) != len(player2.visitedCells) {
		return false
	}
	for k := range player1.visitedCells {
		_, w := player2.visitedCells[k]
		if !w {
			return false
		}
	}
	return true
}

func simulatePlayer(field *Field, limit int, elapsedTurns int, ch chan *Field) *Field {
	turns := 1
	lenTurn := 1
	move := 0
	var movemade int
	for i := 1; i < len(field.Players); i++ {
		if i >= lenTurn {
			turns++
			counter := 1
			for j := lenTurn; j < lenTurn+move; j++ {
				player1 := field.Players[j]
				if player1 != nil {
					for z := j + 1; z < lenTurn+move; z++ {
						player2 := field.Players[z]
						if player2 != nil {
							if playerEqual(player1, player2) {
								for field.Players[lenTurn+move-counter] == nil {
									counter++
								}
								field.Players[z] = field.Players[lenTurn+move-counter]
								field.Players[lenTurn+move-counter] = nil
								counter++
							}
						} else {
							break
						}
					}
				} else {
					break
				}

			}
			counter = 0
			lenTurn = lenTurn + move
			move = 0
		}
		if field.Players[i] != nil {
			probability := 1.0 / math.Pow(5.0, float64(turns))
			for _, action := range Actions {
				field, movemade = simulateMove(field, i, probability, action, elapsedTurns+turns-1, limit)
				move = move + movemade
			}
			field.Players[i] = nil
		} else {
			continue
		}
		if i > limit {
			fmt.Println("Ich habe abgebrochen")
			break
		}
	}
	//	fmt.Println(field.Cells)
	fmt.Println(turns)

	if ch != nil {
		ch <- field
	}
	return field
}

func addFields(field1 *Field, field2 *Field) *Field {
	for i := 0; i < len(field1.Cells); i++ {
		for j := 0; j < len(field1.Cells[i]); j++ {
			field2.Cells[i][j] = field2.Cells[i][j] + field1.Cells[i][j]
		}

	}
	return field2
}

// SpekuClient is a client implementation that uses Minimax to decide what to do next
type SpekuClient struct{}

// GetAction implements the Client interface
//TODO: use player information
func (c SpekuClient) GetAction(player Player, status *Status) Action {

	start := time.Now()
	fieldArray := convertCellsToField(status)
	channels := make(map[int]chan *Field, 0)
	for i, field := range fieldArray {
		if field != nil {
			channels[i] = make(chan *Field)
			go simulatePlayer(field, 100000, status.Turn, channels[i])
		}
	}
	counter := 0
	var targetField *Field
	for _, ch := range channels {
		newField := <-ch
		if counter == 0 {
			targetField = newField
			counter++
		} else {
			targetField = addFields(targetField, newField)
		}

	}
	t := time.Now()

	elapsed := t.Sub(start)
	fmt.Println(elapsed)
	return ChangeNothing
}
