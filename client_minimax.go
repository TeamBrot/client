package main

import (
	"log"
	"math"
	"math/rand"
)

func checkCell(status *Status, direction Direction, y int, x int, fields int, occupiedCells [][]bool) bool {
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
	return status.Cells[y][x] == 0 || (occupiedCells != nil && occupiedCells[y][x])
}

func Moves(status *Status, player *Player, occupiedCells [][]bool) []Action {
	changeNothing := true
	turnRight := true
	turnLeft := true
	slowDown := player.Speed != 1
	speedUp := player.Speed != 10
	for i := 1; i <= player.Speed; i++ {
		checkJump := status.Turn%6 == 0 && i > 1 && i < player.Speed
		checkJumpSlowDown := status.Turn%6 == 0 && i > 1 && i < player.Speed-1
		checkJumpSpeedUp := status.Turn%6 == 0 && i > 1 && i <= player.Speed

		turnLeft = turnLeft && (checkJump || checkCell(status, (player.Direction+1)%4, player.Y, player.X, i, occupiedCells))
		changeNothing = changeNothing && (checkJump || checkCell(status, player.Direction, player.Y, player.X, i, occupiedCells))
		turnRight = turnRight && (checkJump || checkCell(status, (player.Direction+3)%4, player.Y, player.X, i, occupiedCells))
		if i != player.Speed {
			slowDown = slowDown && (checkJumpSlowDown || checkCell(status, player.Direction, player.Y, player.X, i, occupiedCells))
		}
		speedUp = speedUp && (checkJumpSpeedUp || checkCell(status, player.Direction, player.Y, player.X, i, occupiedCells))
	}
	speedUp = speedUp && checkCell(status, player.Direction, player.Y, player.X, player.Speed+1, occupiedCells)

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
	return len(Moves(status, player, nil))
}

// doMove makes the specified player do the specified action, using the specified status.
// An optional occupiedCells can be supplied.
// In this case, a function return value of false indicates that the specified action is valid but would lead to another player dying (the one that created occupiedCells)
// Also, every field newly entered is written into occupiedCells
// The function panics when an illegal move was selected
func doMove(status *Status, playerID int, action Action, occupiedCells [][]bool, writeOccupiedCells bool) bool {
	player := status.Players[playerID]
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
				status.Cells[player.Y][player.X] = playerID
				if writeOccupiedCells {
					occupiedCells[player.Y][player.X] = true
				}
				// defer func() { status.Cells[player.Y][player.X] = 0 }()
			} else if occupiedCells[player.Y][player.X] {
				if playerID == status.You {
					panic("you should never be here")
				}
				return false
			} else {
				log.Println("tried to access", player.Y, player.X, "but field has value", status.Cells[player.Y][player.X])
				panic("this field should always be 0")
			}
		}
	}
	// log.Println("doMove end: ", player.X, player.Y, player.Direction, player.Speed)
	return true

}

func simulate(you int, minimizer int, isMaximizer bool, status *Status, action Action, depth int, alpha int, beta int, ch chan int, occupiedCells [][]bool) int {
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
	//log.Println("Player: ", player)
	if isMaximizer {
		if occupiedCells != nil {
			panic("occupiedCells should be nil if maximizer = true")
		}
		occupiedCells = newOccupiedCells(status)
	} else if occupiedCells == nil {
		panic("occupiedCells should not be nil if maximizer = false")
	}
	// log.Println(depth, "doing move", action, "with speed", status.Players[playerID].Speed, "from", status.Players[playerID].X, status.Players[playerID].Y)
	youSurvived := doMove(status, playerID, action, occupiedCells, isMaximizer)
	if !youSurvived {
		return bestScore
	}
	if depth == 0 && !isMaximizer {
		bestScore = score(status, status.Players[you])
	} else {
		turn := status.Turn
		if isMaximizer {
			m := Moves(status, status.Players[minimizer], occupiedCells)
			// log.Println(depth, "moves for", minimizer, m, depth)
			for _, action := range m {
				sCopy := copyStatus(status)
				score := simulate(you, minimizer, false, sCopy, action, depth, alpha, beta, nil, occupiedCells)

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
			m := Moves(status, status.Players[you], occupiedCells)
			// log.Println(depth, "moves for", you, m, depth)
			for _, action := range m {
				sCopy := copyStatus(status)
				score := simulate(you, minimizer, true, sCopy, action, depth-1, alpha, beta, nil, nil)

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
	if ch != nil {
		ch <- bestScore
	}
	return bestScore
}

func newOccupiedCells(status *Status) [][]bool {
	occupiedCells := make([][]bool, status.Height)
	for i := range occupiedCells {
		occupiedCells[i] = make([]bool, status.Width)
		for j := range occupiedCells[i] {
			occupiedCells[i][j] = false
		}
	}
	return occupiedCells
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
		for j := range s.Cells[i] {
			s.Cells[i][j] = status.Cells[i][j]
		}
	}
	s.Players = make(map[int]*Player)
	for id, player := range status.Players {
		s.Players[id] = &Player{X: player.X, Y: player.Y, Active: player.Active, Name: player.Name, Direction: player.Direction, Speed: player.Speed}
	}
	return &s
}

func findClosestPlayer(status *Status) int {
	ourPlayer := status.Players[status.You]
	nearestPlayer := 0
	nearestPlayerDistance := 0.0
	for playerID, player := range status.Players {
		distance := math.Sqrt(math.Pow(float64(player.X - ourPlayer.X), 2) + math.Pow(float64(player.Y - ourPlayer.Y), 2))
		if playerID != status.You && player.Active && (nearestPlayer == 0 || distance < nearestPlayerDistance) {
			nearestPlayer = playerID
			nearestPlayerDistance = distance
		}
	}
	if nearestPlayer == 0 {
		log.Fatalln("no non-dead player found")
	}
	return nearestPlayer
}

func bestActionsMinimax(maximizerID int, minimizerID int, status *Status, depth int, print bool) []Action {
	bestScore := -100
	bestActions := make([]Action, 0)
	possibleMoves := Moves(status, status.Players[maximizerID], nil)
	// channels := make(map[Action]chan int, 0)
	// for _, action := range possibleMoves {
	// 	channels[action] = make(chan int)
	// 	sCopy := copyStatus(status)
	// 	go simulate(maximizerID, minimizerID, true, sCopy, action, depth, -200, 200, channels[action], nil)
	// }

	for _, action := range possibleMoves {
		sCopy := copyStatus(status)
		score := simulate(maximizerID, minimizerID, true, sCopy, action, depth, -200, 200, nil, nil)
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
func (c MinimaxClient) GetAction(player Player, status *Status) Action {
	actions := bestActionsMinimax(status.You, otherPlayerID, status, 6, true)
	otherPlayerID := findClosestPlayer(status)
	actions := bestActionsMinimax(status.You, otherPlayerID, status, 6)
	if len(actions) == 0 {
		return ChangeNothing
	}
	return actions[rand.Intn(len(actions))]
}
