package main

import (
	"math"
)

// Player contains information on a specific player used by the API
type Player struct {
	X         uint16
	Y         uint16
	Direction Direction
	Speed     uint8
}

// JSONPlayer contains information on a specific player as returned by the server.
type JSONPlayer struct {
	X               int `json:"x"`
	Y               int `json:"y"`
	Direction       Direction
	StringDirection string `json:"direction"`
	Speed           int    `json:"speed"`
	Active          bool   `json:"active"`
	Name            string `json:"name"`
}

// Direction contains the direction the player is facing
type Direction uint8

// turning left is equivalent to +1(mod 4) and turning right to (+3)(mod 4)
const (
	// Up makes the player face up
	Up Direction = iota
	// Left makes the player face left
	Left
	// Down makes the player face down
	Down
	// Right makes the player face right
	Right
)

// Directions maps string direction representation to int representation
var Directions = map[string]Direction{
	"up":    Up,
	"down":  Down,
	"right": Right,
	"left":  Left,
}

// ProcessAction moves the player according to action and turn. Returns visited coordinates
func (player *Player) ProcessAction(action Action, turn uint16) []*Coords {
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
	visitedCoords := make([]*Coords, player.Speed+1)
	jump := turn%6 == 0
	for i := uint8(1); i <= player.Speed; i++ {
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
			visitedCoords[i] = &Coords{player.Y, player.X}
		}
	}
	return visitedCoords
}

// checkCell checks if it is legal for a player to go from a position a certain number of fields
func checkCell(cells [][]bool, direction Direction, y uint16, x uint16, fields uint16, extraCellInfo map[Coords]struct{}, extraCellAllowed bool) bool {
	if direction == Up {
		y -= fields
	} else if direction == Down {
		y += fields
	} else if direction == Left {
		x -= fields
	} else {
		x += fields
	}
	if x >= uint16(len(cells[0])) || y >= uint16(len(cells)) || x < 0 || y < 0 {
		return false
	}
	isPossible := !cells[y][x]
	if extraCellInfo != nil {
		_, fieldVisited := extraCellInfo[Coords{y,x}]
		if extraCellAllowed {
			return isPossible || fieldVisited
		}
		return isPossible && !fieldVisited

	}
	return isPossible
}

// PossibleMoves returns possible actions for a given situation for a player
func (player *Player) PossibleMoves(cells [][]bool, turn uint16, extraCellInfo map[Coords]struct{}, extraCellAllowed bool) []Action {
	changeNothing := true
	turnRight := true
	turnLeft := true
	slowDown := player.Speed != 1
	speedUp := player.Speed != 10
	direction := player.Direction
	y := player.Y
	x := player.X
	for i := uint16(1); i <= uint16(player.Speed); i++ {
		checkJump := turn%6 == 0 && i > 1 && i < uint16(player.Speed)
		checkJumpSlowDown := turn%6 == 0 && i > 1 && i < uint16(player.Speed)-1
		checkJumpSpeedUp := turn%6 == 0 && i > 1 && i <= uint16(player.Speed)

		turnLeft = turnLeft && (checkJump || checkCell(cells, (direction+1)%4, y, x, i, extraCellInfo, extraCellAllowed))
		changeNothing = changeNothing && (checkJump || checkCell(cells, direction, y, x, i, extraCellInfo, extraCellAllowed))
		turnRight = turnRight && (checkJump || checkCell(cells, (direction+3)%4, y, x, i, extraCellInfo, extraCellAllowed))
		if i != uint16(player.Speed) {
			slowDown = slowDown && (checkJumpSlowDown || checkCell(cells, direction, y, x, i, extraCellInfo, extraCellAllowed))
		}
		speedUp = speedUp && (checkJumpSpeedUp || checkCell(cells, direction, y, x, i, extraCellInfo, extraCellAllowed))
	}
	speedUp = speedUp && checkCell(cells, direction, y, x, uint16(player.Speed+1), extraCellInfo, extraCellAllowed)

	possibleMoves := make([]Action, 0, 5)

	if slowDown {
		possibleMoves = append(possibleMoves, SlowDown)
	}
	if changeNothing {
		possibleMoves = append(possibleMoves, ChangeNothing)
	}
	if speedUp {
		possibleMoves = append(possibleMoves, SpeedUp)
	}
	if turnLeft {
		possibleMoves = append(possibleMoves, TurnLeft)
	}
	if turnRight {
		possibleMoves = append(possibleMoves, TurnRight)
	}
	return possibleMoves
}

func (p *Player) DistanceTo(p2 *Player) float64 {
	return math.Sqrt(math.Pow(float64(p.X-p2.X), 2) + math.Pow(float64(p.Y-p2.Y), 2))
}

// ConvertToPlayer converts a JSONPlayer to a Player
func (JSONPlayer *JSONPlayer) ConvertToPlayer() *Player {
	var player Player
	player.X = uint16(JSONPlayer.X)
	player.Y = uint16(JSONPlayer.Y)
	player.Speed = uint8(JSONPlayer.Speed)
	player.Direction = JSONPlayer.Direction
	return &player
}

