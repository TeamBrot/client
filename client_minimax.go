package main

import (
	"errors"
	"log"
	"math"
	"math/rand"
	"time"
)

type MiniMaxResult struct {
	Actions []Action
	Error   error
}

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

func simulate(you int, minimizer int, isMaximizer bool, status *Status, action Action, depth int, alpha int, beta int, occupiedCells [][]bool, stopChannel <-chan time.Time) (int, error) {
	// log.Println("Simulate: ", you, minimizer, action, depth)
	select {
	case <-stopChannel:
		return 0, errors.New("stopped in computation")
	default:
	}

	var playerID int
	var bestScore int
	if isMaximizer {
		playerID = you
		bestScore = 5
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
		return bestScore, nil
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
				score, err := simulate(you, minimizer, false, sCopy, action, depth, alpha, beta, occupiedCells, stopChannel)
				if err != nil {
					return 0, err
				}

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
				score, err := simulate(you, minimizer, true, sCopy, action, depth-1, alpha, beta, nil, stopChannel)
				if err != nil {
					return 0, err
				}

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
	return bestScore, nil
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

func distanceToPlayer(player1 *Player, player2 *Player) float64 {
	return math.Sqrt(math.Pow(float64(player1.X-player2.X), 2) + math.Pow(float64(player1.Y-player2.Y), 2))
}

func findClosestPlayer(status *Status) int {
	ourPlayer := status.Players[status.You]
	nearestPlayer := 0
	nearestPlayerDistance := 0.0
	for playerID, player := range status.Players {
		distance := distanceToPlayer(player, ourPlayer) //math.Sqrt(math.Pow(float64(player.X-ourPlayer.X), 2) + math.Pow(float64(player.Y-ourPlayer.Y), 2))
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

// bestActionsMinimax returns the best actions according to the minimax algorithm.
// it stops execution when a signal is received on the specified channel.
// in this case, the return value should not be used.
func bestActionsMinimax(maximizerID int, minimizerID int, status *Status, depth int, stopChannel <-chan time.Time, returnChannel chan<- MiniMaxResult) ([]Action, error) {
	bestScore := -100
	bestActions := make([]Action, 0)
	possibleMoves := Moves(status, status.Players[maximizerID], nil)

	for _, action := range possibleMoves {
		sCopy := copyStatus(status)
		score, err := simulate(maximizerID, minimizerID, true, sCopy, action, depth, -200, 200, nil, stopChannel)
		if err != nil {
			if returnChannel != nil {
				returnChannel <- MiniMaxResult{Error: err, Actions: []Action{}}
			}
			return []Action{}, err
		}
		if score >= bestScore {
			bestActions = append(bestActions, action)
			bestScore = score
		}
	}
	if len(bestActions) == 0 {
		log.Println("no best actions, using possible moves", possibleMoves)
		if returnChannel != nil {
			returnChannel <- MiniMaxResult{Error: nil, Actions: possibleMoves}
		}
		return possibleMoves, nil
	}
	log.Println("best actions are", bestActions, "with score", bestScore)
	if returnChannel != nil {
		returnChannel <- MiniMaxResult{Error: nil, Actions: bestActions}
	}
	return bestActions, nil
}

func bestActionsMinimaxTimed(maximizerID int, minimizerID int, status *Status, timingChannel <-chan time.Time, resultChannel chan<- []Action) []Action {
	var actions []Action
	startDepth := 4
	maxDepth := 8
	returnChannels := make(map[int]chan MiniMaxResult, 0)
	for depth := startDepth; depth <= maxDepth; depth++ {
		returnChannels[depth] = make(chan MiniMaxResult, 1)
		log.Println("Starting MiniMax with Depth ", depth)
		go bestActionsMinimax(maximizerID, minimizerID, status, depth, timingChannel, returnChannels[depth])
	}

	for reachedDepth, ch := range returnChannels {
		result := <-ch
		if result.Error != nil {
			break
		} else {
			log.Println("Got valid Result for MiniMax with Depth", reachedDepth)
			actions = result.Actions
		}
	}
	resultChannel <- actions
	return actions

}

// MinimaxClient is a client implementation that uses Minimax to decide what to do next
type MinimaxClient struct{}

// GetAction implements the Client interface
func (c MinimaxClient) GetAction(player Player, status *Status, timingChannel <-chan time.Time) Action {
	otherPlayerID := findClosestPlayer(status)
	log.Println("using player", otherPlayerID, "at", status.Players[otherPlayerID].X, status.Players[otherPlayerID].Y, "as minimizer")
	resultChannel := make(chan []Action)
	go bestActionsMinimaxTimed(status.You, otherPlayerID, status, timingChannel, resultChannel)
	actions := <-resultChannel
	if len(actions) == 0 {
		log.Println("no best action, using change_nothing")
		return ChangeNothing
	}
	action := actions[rand.Intn(len(actions))]
	log.Println("multiple best actions, using", action)
	return action
}
