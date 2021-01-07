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
		player.Direction = (player.Direction + 1) % 4
	} else if action == TurnRight {
		player.Direction = (player.Direction + 3) % 4
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
	intY := int(y)
	intX := int(x)
	if direction == Up {
		intY = int(y) - int(fields)
	} else if direction == Down {
		intY = int(y) + int(fields)
	} else if direction == Left {
		intX = int(x) - int(fields)
	} else {
		intX = int(x) + int(fields)
	}
	if intX >= len(cells[0]) || intY >= len(cells) || intX < 0 || intY < 0 {
		return false
	}
	isPossible := !cells[intY][intX]
	if extraCellInfo != nil {
		_, fieldVisited := extraCellInfo[Coords{uint16(intY), uint16(intX)}]
		if extraCellAllowed {
			return isPossible || fieldVisited
		}
		return isPossible && !fieldVisited

	}
	return isPossible
}

// PossibleActions returns possible actions for a given situation for a player
func (player *Player) PossibleActions(cells [][]bool, turn uint16, extraCellInfo map[Coords]struct{}, extraCellAllowed bool) []Action {
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

	possibleActions := make([]Action, 0, 5)

	if changeNothing {
		possibleActions = append(possibleActions, ChangeNothing)
	}
	if turnLeft {
		possibleActions = append(possibleActions, TurnLeft)
	}
	if turnRight {
		possibleActions = append(possibleActions, TurnRight)
	}
	if speedUp {
		possibleActions = append(possibleActions, SpeedUp)
	}
	if slowDown {
		possibleActions = append(possibleActions, SlowDown)
	}
	return possibleActions
}

//Returns the distance between to players as float64
func (player *Player) DistanceTo(p2 *Player) float64 {
	return math.Sqrt(math.Pow(float64(int(player.X)-int(p2.X)), 2) + math.Pow(float64(int(player.Y)-int(p2.Y)), 2))
}

//This function copies a struct of type Player
func (player *Player) Copy() *Player {
	newPlayer := *player
	return &newPlayer
}
