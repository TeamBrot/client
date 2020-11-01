package main

import "log"

func checkCell(status *Status, current bool, y int, x int, jump bool) bool {
	if x >= status.Width || y >= status.Height || x < 0 || y < 0 {
		return false
	}
	return (status.Cells[y][x] == 0 || jump) && current
}

func moves(status *Status, player *Player) []Action {
	changeNothing := true
	turnRight := true
	turnLeft := true
	speedDown := true
	speedUp := true
	for i := 1; i <= player.Speed; i++ {
		checkJump := status.Turn%6 == 0 && i > 1 && i < player.Speed
		checkJumpSpeedDown := status.Turn%6 == 0 && i > 1 && i < player.Speed-1
		checkJumpSpeedUp := status.Turn%6 == 0 && i > 1 && i <= player.Speed
		if player.Direction == Right {
			turnRight = checkCell(status, turnRight, player.Y+i, player.X, checkJump)
			changeNothing = checkCell(status, changeNothing, player.Y, player.X+i, checkJump)
			if i != player.Speed {
				speedDown = checkCell(status, speedDown, player.Y, player.X+i, checkJumpSpeedDown)
			}
			speedUp = checkCell(status, speedUp, player.Y, player.X+i, checkJumpSpeedUp)
			turnLeft = checkCell(status, turnLeft, player.Y-i, player.X, checkJump)
		} else if player.Direction == Up {
			turnRight = checkCell(status, turnRight, player.Y, player.X+i, checkJump)
			changeNothing = checkCell(status, changeNothing, player.Y-i, player.X, checkJump)
			if i != player.Speed {
				speedDown = checkCell(status, speedDown, player.Y-i, player.X, checkJumpSpeedDown)
			}
			speedUp = checkCell(status, speedUp, player.Y-i, player.X, checkJumpSpeedUp)
			turnLeft = checkCell(status, turnLeft, player.Y, player.X-i, checkJump)
		} else if player.Direction == Left {
			turnRight = checkCell(status, turnRight, player.Y-i, player.X, checkJump)
			changeNothing = checkCell(status, changeNothing, player.Y, player.X-i, checkJump)
			if i != player.Speed {
				speedDown = checkCell(status, speedDown, player.Y, player.X-i, checkJumpSpeedDown)
			}
			speedUp = checkCell(status, speedUp, player.Y, player.X-i, checkJumpSpeedUp)
			turnLeft = checkCell(status, turnLeft, player.Y+i, player.X, checkJump)
		} else if player.Direction == Down {
			turnRight = checkCell(status, turnRight, player.Y, player.X-i, checkJump)
			changeNothing = checkCell(status, changeNothing, player.Y+i, player.X, checkJump)
			if i != player.Speed {
				speedDown = checkCell(status, speedDown, player.Y+i, player.X, checkJumpSpeedDown)
			}
			speedUp = checkCell(status, speedUp, player.Y+i, player.X, checkJumpSpeedUp)
			turnLeft = checkCell(status, turnLeft, player.Y, player.X+i, checkJump)
		}
	}

	if player.Direction == Right {
		speedUp = checkCell(status, speedUp, player.Y, player.X+player.Speed+1, false)
	} else if player.Direction == Up {
		speedUp = checkCell(status, speedUp, player.Y-player.Speed-1, player.X, false)
	} else if player.Direction == Left {
		speedUp = checkCell(status, speedUp, player.Y, player.X-player.Speed-1, false)
	} else if player.Direction == Down {
		speedUp = checkCell(status, speedUp, player.Y+player.Speed+1, player.X, false)
	}

	possibleMoves := make([]Action, 0)

	if speedDown && player.Speed != 1 {
		possibleMoves = append(possibleMoves, SlowDown)
	}
	if changeNothing {
		possibleMoves = append(possibleMoves, ChangeNothing)
	}
	if speedUp && player.Speed != 10 {
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

func score(status *Status, player *Player) int {
	return len(moves(status, player))
}

func doMove(status *Status, player *Player, action Action) {
	log.Println("doMove start: ", player.X, player.Y, player.Direction, player.Speed)
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
	jump := status.Turn%6 == 0
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
			if status.Cells[player.Y][player.X] == 0 {
				status.Cells[player.Y][player.X] = status.You
				// defer func() { status.Cells[player.Y][player.X] = 0 }()
			} else {
				panic("the field should always be 0")
			}
		}
	}
	log.Println("doMove end: ", player.X, player.Y, player.Direction, player.Speed)

}

func undoMove(status *Status, player *Player, action Action) {
	log.Println("undoMove start: ", player.X, player.Y, player.Direction, player.Speed)
	jump := status.Turn%6 == 0
	for i := 1; i <= player.Speed; i++ {
		if !jump || i == 1 || i == player.Speed {
			status.Cells[player.Y][player.X] = 0
		}

		if player.Direction == Up {
			player.Y++
		} else if player.Direction == Down {
			player.Y--
		} else if player.Direction == Right {
			player.X--
		} else if player.Direction == Left {
			player.X++
		}
	}

	//Asserting that speedUp was not used at speed 10 and slowDown not at 1
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
	log.Println("undoMove end: ", player.X, player.Y, player.Direction, player.Speed)
}

func simulate(you int, remainingPlayers []int, status *Status, action Action, depth int) int {
	log.Println("Simulate: ", you, remainingPlayers, action, depth)
	if len(remainingPlayers) == 0 {
		panic("There should always be remaining players")
	}

	playerID := remainingPlayers[0]
	player := status.Players[playerID]
	log.Println("Player: ", player)
	remainingPlayers = remainingPlayers[1:]
	doMove(status, player, action)
	var bestScore int
	if playerID == you {
		bestScore = -100
	} else {
		bestScore = 100
	}
	if len(remainingPlayers) == 0 && depth == 0 {
		bestScore = score(status, status.Players[you])

	} else {
		turn := status.Turn
		if len(remainingPlayers) == 0 {
			for id, player := range status.Players {
				if player.Active {
					remainingPlayers = append(remainingPlayers, id)
				}
			}
			depth--
			status.Turn++
		}
		for _, action := range moves(status, status.Players[remainingPlayers[0]]) {
			score := simulate(you, remainingPlayers, status, action, depth)

			if score > bestScore && playerID == you || score < bestScore && playerID != you {
				bestScore = score
			}
		}
		status.Turn = turn
	}
	undoMove(status, player, action)
	return bestScore
}

// MinimaxClient is a client implementation that uses Minimax to decide what to do next
type MinimaxClient struct{}

// GetAction implements the Client interface
//TODO: use player information
func (c MinimaxClient) GetAction(player Player, status *Status) Action {
	remainingPlayers := make([]int, 1)
	remainingPlayers[0] = status.You
	for id, player := range status.Players {
		if player.Active && status.You != id {
			remainingPlayers = append(remainingPlayers, id)
		}
	}
	bestScore := -1
	var bestAction Action
	for _, action := range moves(status, status.Players[status.You]) {
		score := simulate(status.You, remainingPlayers, status, action, 2)
		if score > bestScore {
			bestAction = action
			bestScore = score
		}
	}
	log.Println("bestAction: ", bestAction, bestScore)
	return bestAction
}
