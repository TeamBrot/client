package main

import (
	"errors"
	"log"
	"math"
	"math/rand"
	"time"
)

func score(status *Status, player *Player) int {
	return len(possibleMoves(player, status.Cells, status.Turn, nil, false))
}

// doMove makes the specified player do the specified action, using the specified status.
// An optional occupiedCells can be supplied.
// In this case, a function return value of false indicates that the specified action is valid but would lead to another player dying (the one that created occupiedCells)
// Also, every field newly entered is written into occupiedCells
// The function panics when an illegal move was selected
func doMove(status *Status, playerID uint8, action Action, occupiedCells map[Coords]struct{}, writeOccupiedCells bool) bool {
	player := status.Players[playerID]
	// log.Println("doMove start: ", player.X, player.Y, player.Direction, player.Speed)
	visitedCoords := player.processAction(action, status.Turn)
	for _, coords := range visitedCoords {
		if coords == nil {
			continue
		}
		_, isIn := occupiedCells[*coords]

		if !status.Cells[coords.Y][coords.X] {
			status.Cells[coords.Y][coords.X] = true
			if writeOccupiedCells {
				occupiedCells[*coords] = struct{}{}
			}
			// defer func() { status.Cells[player.Y][player.X] = 0 }()
		} else if isIn {
			if playerID == status.You {
				panic("you should never be here")
			}
			return false
		} else {
			log.Println("tried to access", player.Y, player.X, "but field has value", status.Cells[player.Y][player.X])
			panic("this field should always be false")
		}

	}
	// log.Println("doMove end: ", player.X, player.Y, player.Direction, player.Speed)
	return true

}

func simulate(you uint8, minimizer uint8, isMaximizer bool, status *Status, action Action, depth int, alpha int, beta int, occupiedCells map[Coords]struct{}, stopChannel <-chan time.Time) (int, error) {
	// log.Println("Simulate: ", you, minimizer, action, depth)
	select {
	case <-stopChannel:
		return 0, errors.New("stopped in computation")
	default:
	}

	var playerID uint8
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
		occupiedCells = make(map[Coords]struct{})
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
			m := possibleMoves(status.Players[minimizer], status.Cells, status.Turn, occupiedCells, true)
			// log.Println(depth, "moves for", minimizer, m, depth)
			for _, action := range m {
				sCopy := status.copyStatus()
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
			m := possibleMoves(status.Players[you], status.Cells, status.Turn, occupiedCells, true)
			// log.Println(depth, "moves for", you, m, depth)
			for _, action := range m {
				sCopy := status.copyStatus()
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

func (status *Status) copyStatus() *Status {
	var s Status
	s.Width = status.Width
	s.Height = status.Height
	s.Turn = status.Turn
	s.You = status.You
	s.Cells = make([][]bool, s.Height)
	for i := range s.Cells {
		s.Cells[i] = make([]bool, status.Width)
		for j := range s.Cells[i] {
			s.Cells[i][j] = status.Cells[i][j]
		}
	}
	s.Players = make(map[uint8]*Player)
	for id, player := range status.Players {
		s.Players[id] = player.copyPlayer()
	}
	return &s
}

func distanceToPlayer(player1 *Player, player2 *Player) float64 {
	return math.Sqrt(math.Pow(float64(player1.X-player2.X), 2) + math.Pow(float64(player1.Y-player2.Y), 2))
}

func findClosestPlayer(status *Status) uint8 {
	ourPlayer := status.Players[status.You]
	var nearestPlayer uint8
	nearestPlayerDistance := 0.0
	for playerID, player := range status.Players {
		distance := distanceToPlayer(player, ourPlayer) //math.Sqrt(math.Pow(float64(player.X-ourPlayer.X), 2) + math.Pow(float64(player.Y-ourPlayer.Y), 2))
		if playerID != status.You && (nearestPlayer == 0 || distance < nearestPlayerDistance) {
			nearestPlayer = playerID
			nearestPlayerDistance = distance
		}
	}
	if nearestPlayer == 0 {
		log.Fatalln("no non-dead player found")
	}
	return nearestPlayer
}

func minimaxTiming(calculationTime time.Duration, timingChannel chan<- time.Time) {
	time.Sleep(time.Duration(0.9 * float64(calculationTime.Nanoseconds())))
	close(timingChannel)
}

// bestActionsMinimax returns the best actions according to the minimax algorithm.
// it stops execution when a signal is received on the specified channel.
// in this case, the return value should not be used.
func bestActionsMinimax(maximizerID uint8, minimizerID uint8, status *Status, depth int, stopChannel <-chan time.Time) ([]Action, error) {
	bestScore := -100
	bestActions := make([]Action, 0)
	possibleMoves := possibleMoves(status.Players[maximizerID], status.Cells, status.Turn, nil, true)
	for _, action := range possibleMoves {
		sCopy := status.copyStatus()
		score, err := simulate(maximizerID, minimizerID, true, sCopy, action, depth, -200, 200, nil, stopChannel)
		if err != nil {
			return []Action{}, err
		}
		if score >= bestScore {
			bestActions = append(bestActions, action)
			bestScore = score
		}
	}
	if len(bestActions) == 0 {
		log.Println("no best actions, using possible moves", possibleMoves)
		return possibleMoves, nil
	}
	log.Println("best actions are", bestActions, "with score", bestScore)
	return bestActions, nil
}

func bestActionsMinimaxTimed(maximizerID uint8, minimizerID uint8, status *Status, timingChannel <-chan time.Time) []Action {
	var actions []Action
	var depth int
	startDepth := 1
	depth += startDepth
	for {
		//Backup in Case we can not finish in time with startDepth
		actions = possibleMoves(status.Players[status.You], status.Cells, status.Turn, nil, true)
		if len(actions) == 0 {
			return []Action{ChangeNothing}
		} else if len(actions) == 1 {
			return actions
		}
		actionsTemp, err := bestActionsMinimax(maximizerID, minimizerID, status, depth, timingChannel)
		if err == nil {
			log.Println("minimax with depth", depth, "actions", actionsTemp, "no error")
			actions = actionsTemp
		} else {
			log.Println("couldn't finish calculation for depth", depth)
			return actions
		}
		depth++
	}
}

// MinimaxClient is a client implementation that uses Minimax to decide what to do next
type MinimaxClient struct{}

// GetAction implements the Client interface
func (c MinimaxClient) GetAction(player Player, status *Status, calculationTime time.Duration) Action {
	stopChannel := make(chan time.Time)
	go minimaxTiming(calculationTime, stopChannel)
	otherPlayerID := findClosestPlayer(status)
	log.Println("using player", otherPlayerID, "at", status.Players[otherPlayerID].X, status.Players[otherPlayerID].Y, "as minimizer")
	actions := bestActionsMinimaxTimed(status.You, otherPlayerID, status, stopChannel)
	if len(actions) == 0 {
		log.Println("no best action, using change_nothing")
		return ChangeNothing
	}
	action := actions[rand.Intn(len(actions))]
	log.Println("multiple best actions, using", action)
	return action
}
