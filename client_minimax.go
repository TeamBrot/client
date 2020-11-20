package main

import (
	"log"
	"math/rand"
)

func checkCell(status *Status, direction Direction, y int, x int, fields int) bool {
	if direction == Up {
		y -= fields
	} else if direction == Down {
		y += fields
	} else if direction == Left {
		x -= fields
	} else {
		x += fields
	}
	if x >= status.Width || y >= status.Height || x < 0 || y < 0 {
		return false
	}
	return status.Cells[y][x] == 0
}

func moves(status *Status, player *Player) []Action {
	changeNothing := true
	turnRight := true
	turnLeft := true
	slowDown := player.Speed != 1
	speedUp := player.Speed != 10
	for i := 1; i <= player.Speed; i++ {
		checkJump := status.Turn%6 == 0 && i > 1 && i < player.Speed
		checkJumpSlowDown := status.Turn%6 == 0 && i > 1 && i < player.Speed-1
		checkJumpSpeedUp := status.Turn%6 == 0 && i > 1 && i <= player.Speed

		turnLeft = turnLeft && (checkJump || checkCell(status, (player.Direction+1)%4, player.Y, player.X, i))
		changeNothing = changeNothing && (checkJump || checkCell(status, player.Direction, player.Y, player.X, i))
		turnRight = turnRight && (checkJump || checkCell(status, (player.Direction+3)%4, player.Y, player.X, i))
		if i != player.Speed {
			slowDown = slowDown && (checkJumpSlowDown || checkCell(status, player.Direction, player.Y, player.X, i))
		}
		speedUp = speedUp && (checkJumpSpeedUp || checkCell(status, player.Direction, player.Y, player.X, i))
	}
	speedUp = speedUp && checkCell(status, player.Direction, player.Y, player.X, player.Speed+1)

	possibleMoves := make([]Action, 0)

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
		s.Cells[i] = make([]int, status.Width)
		copy(s.Cells[i], status.Cells[i])
	}
	s.Players = make(map[int]*Player)
	for id, player := range status.Players {
		s.Players[id] = &Player{X: player.X, Y: player.Y, Active: player.Active, Name: player.Name, Direction: player.Direction, Speed: player.Speed}
	}
	return &s
}

func findClosestPlayer(playerID int, status *Status) int {
	//TODO: write function
	return 0
}

func bestActionsMinimax(maximizerID int, minimizerID int, status *Status, depth int, print bool) []Action {
	bestScore := -100
	bestActions := make([]Action, 0)
	possibleMoves := moves(status, status.Players[maximizerID])
	channels := make(map[Action]chan int, 0)
	for _, action := range possibleMoves {
		channels[action] = make(chan int)
		sCopy := copyStatus(status)
		go simulate(maximizerID, minimizerID, true, sCopy, action, depth, -200, 200, channels[action])
	}

	for action, ch := range channels {
		score := <-ch
		if score >= bestScore {
			bestActions = append(bestActions, action)
			bestScore = score
		}
	}
	if len(bestActions) == 0 {
		if print {
			log.Println("No best action, possibleMoves: ", possibleMoves)
		}
		return possibleMoves
	}
	if print {
		log.Println("bestActions: ", bestActions, bestScore)
	}
	return bestActions
}

// MinimaxClient is a client implementation that uses Minimax to decide what to do next
type MinimaxClient struct{}

// GetAction implements the Client interface
//TODO: use player information
func (c MinimaxClient) GetAction(player Player, status *Status) Action {
	//change remainingPlayers to minimizer and choose minimizer from active remaining players
	// TODO: use closest player
	// TODO: make player move at the same time
	// findClosestPlayer(status.You, status)
	var otherPlayerID int
	for id, player := range status.Players {
		if player.Active && status.You != id {
			otherPlayerID = id
		}
	}
	actions := bestActionsMinimax(status.You, otherPlayerID, status, 7, true)
	if len(actions) == 0 {
		return ChangeNothing
	}
	return actions[rand.Intn(len(actions))]
}
