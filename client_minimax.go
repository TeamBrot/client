package main

import (
	"log"
	"math/rand"
)

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
	// log.Println("doMove start: ", player.X, player.Y, player.Direction, player.Speed)
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
	// log.Println("doMove end: ", player.X, player.Y, player.Direction, player.Speed)

}

func undoMove(status *Status, player *Player, action Action) {
	// log.Println("undoMove start: ", player.X, player.Y, player.Direction, player.Speed)
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
	// log.Println("undoMove end: ", player.X, player.Y, player.Direction, player.Speed)
}

func simulate(you int, minimizer int, isMaximizer bool, status *Status, action Action, depth int, alpha int, beta int, ch chan int) int {
	// log.Println("Simulate: ", you, minimizer, action, depth)

	var playerID int
	var bestScore int
	if isMaximizer {
		playerID = you
		bestScore = 100 + depth
	} else {
		playerID = minimizer
		bestScore = -100 - depth
	}
	player := status.Players[playerID]
	//log.Println("Player: ", player)
	doMove(status, player, action)
	if depth == 0 && !isMaximizer {
		bestScore = score(status, status.Players[you])
	} else {
		turn := status.Turn
		if isMaximizer {
			for _, action := range moves(status, status.Players[minimizer]) {
				score := simulate(you, minimizer, false, status, action, depth, alpha, beta, nil)

				if score < bestScore {
					bestScore = score
				}
				if bestScore < beta {
					beta = bestScore
				}

				if beta <= alpha {
					break
				}
			}
		} else {
			status.Turn++
			for _, action := range moves(status, status.Players[you]) {
				score := simulate(you, minimizer, true, status, action, depth-1, alpha, beta, nil)

				if score > bestScore {
					bestScore = score
				}
				if bestScore > alpha {
					alpha = bestScore
				}

				if beta <= alpha {
					break
				}
			}
		}
		status.Turn = turn
	}
	undoMove(status, player, action)
	if ch != nil {
		ch <- bestScore
	}
	return bestScore
}

func copyStatus(status *Status) *Status {
	var s Status
	s.Width = status.Width
	s.Height = status.Height
	s.Deadline = status.Deadline
	s.Running = status.Running
	s.Turn = status.Turn
	s.You = status.You
	s.Cells = make([][]int, s.Height)
	for i := range s.Cells {
		s.Cells[i] = make([]int, status.Height)
		copy(s.Cells[i], status.Cells[i])
	}
	s.Players = make(map[int]*Player)
	for id, player := range status.Players {
		s.Players[id] = &Player{X: player.X, Y: player.Y, Active: player.Active, Name: player.Name, Direction: player.Direction, Speed: player.Speed}
	}
	return &s
}

// MinimaxClient is a client implementation that uses Minimax to decide what to do next
type MinimaxClient struct{}

// GetAction implements the Client interface
//TODO: use player information
func (c MinimaxClient) GetAction(player Player, status *Status) Action {
	//change remainingPlayers to minimizer and choose minimizer from active remaining players
	// TODO: use closest player
	var otherPlayer int
	for id, player := range status.Players {
		if player.Active && status.You != id {
			otherPlayer = id
		}
	}
	bestScore := -100
	bestActions := make([]Action, 0)
	possibleMoves := moves(status, status.Players[status.You])
	channels := make(map[Action]chan int, 0)
	for _, action := range possibleMoves {
		channels[action] = make(chan int)
		sCopy := copyStatus(status)
		go simulate(status.You, otherPlayer, true, sCopy, action, 7, -200, 200, channels[action])
	}

	for action, ch := range channels {
		score := <-ch
		if score >= bestScore {
			bestActions = append(bestActions, action)
			bestScore = score
		}
	}
	log.Println("bestActions: ", bestActions, bestScore)
	if len(bestActions) == 0 {
		if len(possibleMoves) == 0 {
			return "change_nothing"
		}
		bestActions = possibleMoves
	}
	return bestActions[rand.Intn(len(bestActions))]
}
