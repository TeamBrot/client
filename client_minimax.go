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

func simulate(player Player, status *Status, action Action) int {
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
	score := len(moves(status, &player))
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

	return score
}

// MinimaxClient is a client implementation that uses Minimax to decide what to do next
type MinimaxClient struct{}

// GetAction implements the Client interface
func (c MinimaxClient) GetAction(player Player, status *Status) Action {
	var bestAction Action
	bestScore := -1
	for _, action := range moves(status, status.Players[status.You]) {
		score := simulate(*status.Players[status.You], status, action)
		if score > bestScore {
			bestAction = action
			bestScore = score
		}
		log.Println("Looked at: ", action, score)
	}
	log.Println("bestAction: ", bestAction, bestScore)
	return bestAction
}
