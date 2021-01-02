package main

import (
	"errors"
	"log"
	"math/rand"
	"time"
)

func score(status *Status, player *Player) int {
	return len(player.PossibleActions(status.Cells, status.Turn, nil, false))
}

// doMove makes the specified player do the specified action, using the specified status.
// An optional occupiedCells can be supplied.
// In this case, a function return value of false indicates that the specified action is valid but would lead to another player dying (the one that created occupiedCells)
// Also, every field newly entered is written into occupiedCells
// The function panics when an illegal move was selected
func doMove(status *Status, playerID uint8, action Action, occupiedCells map[Coords]struct{}, writeOccupiedCells bool) bool {
	player := status.Players[playerID]
	// log.Println("doMove start: ", player.X, player.Y, player.Direction, player.Speed)
	visitedCoords := player.ProcessAction(action, status.Turn)
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
		} else if isIn {
			if playerID == status.You {
				panic("you should never be here")
			}
			return false
		} else {
			log.Println("tried to access", player.Y, player.X, "but field is already true")
			panic("this field should always be false")
		}

	}
	// log.Println("doMove end: ", player.X, player.Y, player.Direction, player.Speed)
	return true

}

func getActionScore(you uint8, minimizer uint8, isMaximizer bool, status *Status, action Action, depth int, alpha int, beta int, occupiedCells map[Coords]struct{}, stopChannel <-chan time.Time) (int, int, error) {
	// log.Println("Simulate: ", you, minimizer, action, depth)
	select {
	case <-stopChannel:
		return 0, 0, errors.New("stopped in computation")
	default:
	}

	var playerID uint8
	var bestScore int
	if isMaximizer {
		playerID = you
		bestScore = 5
	} else {
		playerID = minimizer
		bestScore = -1 - depth
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
		return bestScore, 0, nil
	}
	maxDepth := 0
	if depth == 0 && !isMaximizer {
		bestScore = score(status, status.Players[you])
	} else {
		turn := status.Turn
		if isMaximizer {
			m := status.Players[minimizer].PossibleActions(status.Cells, status.Turn, occupiedCells, true)
			// log.Println(depth, "moves for", minimizer, m, depth)
			for _, action := range m {
				sCopy := status.Copy()
				score, d, err := getActionScore(you, minimizer, false, sCopy, action, depth, alpha, beta, occupiedCells, stopChannel)
				if err != nil {
					return 0, d, err
				}

				if score < bestScore {
					bestScore = score
				}
				if d > maxDepth {
					maxDepth = d
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
			m := status.Players[you].PossibleActions(status.Cells, status.Turn, occupiedCells, true)
			// log.Println(depth, "moves for", you, m, depth)
			for _, action := range m {
				sCopy := status.Copy()
				score, d, err := getActionScore(you, minimizer, true, sCopy, action, depth-1, alpha, beta, nil, stopChannel)
				if err != nil {
					return 0, d + 1, err
				}

				if score > bestScore {
					bestScore = score
				}
				if d+1 > maxDepth {
					maxDepth = d + 1
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
	return bestScore, maxDepth, nil
}

// MinimaxScoreMap returns a map that contains the score of every possible move, given a certain depth.
// It could also be empty, indicating that there are no possible moves.
// It stops execution when a signal is received on the specified channel.
// In this case, the return value should not be used.
func MinimaxScoreMap(maximizerID uint8, minimizerID uint8, status *Status, depth int, stopChannel <-chan time.Time) (map[Action]int, int, error) {
	scoreMap := map[Action]int{}
	possibleMoves := status.Players[maximizerID].PossibleActions(status.Cells, status.Turn, nil, true)
	maxDepth := 0
	for _, action := range possibleMoves {
		sCopy := status.Copy()
		score, depth, err := getActionScore(maximizerID, minimizerID, true, sCopy, action, depth, -200, 200, nil, stopChannel)
		if err != nil {
			return map[Action]int{}, maxDepth, err
		}
		if depth > maxDepth {
			maxDepth = depth
		}
		scoreMap[action] = score
	}
	return scoreMap, maxDepth, nil
}

// MinimaxBestActions returns the best actions according to the minimax algorithm, with given depth.
// It stops execution when a signal is received on the specified channel.
// In this case, the return value should not be used.
func MinimaxBestActions(maximizerID uint8, minimizerID uint8, status *Status, depth int, stopChannel <-chan time.Time) ([]Action, int, error) {
	scoreMap, maxDepth, err := MinimaxScoreMap(maximizerID, minimizerID, status, depth, stopChannel)
	if err != nil {
		return []Action{}, maxDepth, err
	}
	if len(scoreMap) == 0 {
		log.Println("no possible moves, using change_nothing")
		return []Action{ChangeNothing}, maxDepth, nil
	}
	bestActions := []Action{}
	bestScore := -1000
	for action, score := range scoreMap {
		if score == bestScore {
			bestActions = append(bestActions, action)
		} else if score > bestScore {
			bestActions = []Action{action}
			bestScore = score
		}
	}
	log.Println("action-score-map", scoreMap)
	return bestActions, maxDepth, nil
}

// MinimaxBestActionsTimed returns the best actions according to the minimax algorithm.
// It stops execution when a signal is received on the specified channel.
// In this case, the return value is the best one available.
func MinimaxBestActionsTimed(maximizerID uint8, minimizerID uint8, status *Status, timingChannel <-chan time.Time) []Action {
	var actions []Action
	var depth int
	startDepth := 1
	depth += startDepth

	actions = status.Players[status.You].PossibleActions(status.Cells, status.Turn, nil, true)
	if len(actions) == 0 {
		return []Action{ChangeNothing}
	} else if len(actions) == 1 {
		return actions
	}
	for {
		sCopy := status.Copy()
		actionsTemp, maxDepth, err := MinimaxBestActions(maximizerID, minimizerID, sCopy, depth, timingChannel)
		if err == nil {
			log.Println("minimax with depth", depth, "actions", actionsTemp)
			actions = actionsTemp
		} else {
			log.Println("couldn't finish calculation for depth", depth, "returning", actions)
			return actions
		}
		if depth > maxDepth {
			log.Println("maximum depth was", maxDepth, "returning", actionsTemp)
			return actions
		}
		depth++
	}
}

func minimaxBestActionsMultiplePlayers(otherPlayerIDs []uint8, myID uint8, status *Status, stopChannel chan time.Time) []Action {
	resultChannels := make(map[uint8]chan []Action)
	for _, otherPlayerID := range otherPlayerIDs {
		resultChannels[otherPlayerID] = make(chan []Action)
		log.Println("using player", otherPlayerID, "at", status.Players[otherPlayerID].X, status.Players[otherPlayerID].Y, "as minimizer")
		go func(otherPlayerID uint8) {
			bestActionsMinimax := MinimaxBestActionsTimed(status.You, otherPlayerID, status, stopChannel)
			resultChannels[otherPlayerID] <- bestActionsMinimax
		}(otherPlayerID)
	}
	allMiniMaxActions := make([]map[Action]struct{}, len(otherPlayerIDs))
	counter := 0
	for _, channel := range resultChannels {
		minimaxActions := <-channel
		actionStruct := make(map[Action]struct{}, 0)
		for _, action := range minimaxActions {
			actionStruct[action] = struct{}{}
		}
		allMiniMaxActions[counter] = actionStruct
		counter++
	}
	bestActionsMinimax := make([]Action, 0)
	for _, action := range Actions {
		for z, actionMap := range allMiniMaxActions {
			_, isIn := actionMap[action]
			if !isIn {
				break
			}
			if z+1 == len(allMiniMaxActions) {
				bestActionsMinimax = append(bestActionsMinimax, action)
			}
		}
	}
	log.Println(bestActionsMinimax)
	return bestActionsMinimax
}

// MinimaxClient is a client implementation that uses Minimax to decide what to do next
type MinimaxClient struct{}

// GetAction implements the Client interface
func (c MinimaxClient) GetAction(player Player, status *Status, calculationTime time.Duration) Action {
	stopChannel := time.After((calculationTime / 10) * 9)
	otherPlayerID, err := status.FindClosestPlayerTo(status.You)
	if err != nil {
		log.Println("could not find closest player:", err)
		return ChangeNothing
	}
	log.Println("using player", otherPlayerID, "at", status.Players[otherPlayerID].X, status.Players[otherPlayerID].Y, "as minimizer")
	actions := MinimaxBestActionsTimed(status.You, otherPlayerID, status, stopChannel)
	if len(actions) > 0 {
		return actions[rand.Intn(len(actions))]
	}
	log.Println("no best action, using change_nothing")
	return ChangeNothing
}
