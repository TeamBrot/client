package main

import (
	"errors"
	"log"
	"math"
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

func combineScoreMaps(scoreMaps []map[Action]int) map[Action]int {
	resultScoreMap := make(map[Action]int, 0)
	for _, action := range Actions {
		minimumScore := math.MaxInt32
		actionPossible := true
		for _, scoreMap := range scoreMaps {
			score, isIn := scoreMap[action]
			if !isIn {
				actionPossible = false
				break
			}
			if score < minimumScore {
				minimumScore = score
			}

		}
		if actionPossible {
			resultScoreMap[action] = minimumScore
		}
	}
	return resultScoreMap
}

func bestActionsFromScoreMap(scoreMap map[Action]int) []Action {
	bestActions := []Action{}
	bestScore := math.MinInt32
	for action, score := range scoreMap {
		log.Println(action, score)
		if score > bestScore {
			bestScore = score
			bestActions = []Action{action}
		} else if score == bestScore {
			bestActions = append(bestActions, action)
		}
	}
	return bestActions
}

func getScoreMapDepth(maximizerID uint8, minimizerID uint8, status *Status, depth int, stopChannel <-chan time.Time) (map[Action]int, int, error) {
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

// getScoreMap returns a score map that contains the score that can be reached against the specified player other players
func getScoreMap(maximizerID uint8, minimizerID uint8, status *Status, timingChannel <-chan time.Time) map[Action]int {
	var scoreMap map[Action]int
	var depth int
	startDepth := 1
	depth += startDepth

	for {
		sCopy := status.Copy()
		scoreMapTemp, maxDepth, err := getScoreMapDepth(maximizerID, minimizerID, sCopy, depth, timingChannel)
		if err == nil {
			scoreMap = scoreMapTemp
			log.Println("minimax with depth", depth, "score map", scoreMap)
		} else {
			log.Println("minimax couldn't finish calculation for depth", depth, "returning", scoreMap)
			return scoreMap
		}
		if depth > maxDepth {
			log.Println("minimax maximum depth was", maxDepth, "returning", scoreMap)
			return scoreMap
		}
		depth++
	}
}

// getScoreMapMultiplePlayers returns a score map that contains the minimum score that can be reached against all other players
// the stop channel has to be closed when the process should stop
func getScoreMapMultiplePlayers(maximizerID uint8, otherPlayerIDs []uint8, status *Status, stopChannel <-chan time.Time) map[Action]int {
	resultChannels := make(map[uint8]chan map[Action]int)
	for _, otherPlayerID := range otherPlayerIDs {
		resultChannels[otherPlayerID] = make(chan map[Action]int)
		log.Println("using player", otherPlayerID, "at", status.Players[otherPlayerID].X, status.Players[otherPlayerID].Y, "as minimizer")
		go func(id uint8) {
			scoreMap := getScoreMap(status.You, id, status, stopChannel)
			resultChannels[id] <- scoreMap
		}(otherPlayerID)
	}
	scoreMaps := make([]map[Action]int, len(otherPlayerIDs))
	i := 0
	for _, ch := range resultChannels {
		scoreMaps[i] = <-ch
		i++
	}
	return combineScoreMaps(scoreMaps)
}

// MinimaxBestActions returns the best actions according to the minimax algorithm
// It stops execution when the specified channel is closed
func MinimaxBestActions(maximizerID uint8, otherPlayerIDs []uint8, status *Status, stopChannel <-chan time.Time) []Action {
	actions := status.Players[status.You].PossibleActions(status.Cells, status.Turn, nil, true)
	if len(actions) == 0 {
		return []Action{ChangeNothing}
	} else if len(actions) == 1 {
		return actions
	}

	if len(otherPlayerIDs) == 1 {
		return bestActionsFromScoreMap(getScoreMap(maximizerID, otherPlayerIDs[0], status, stopChannel))
	}
	return bestActionsFromScoreMap(getScoreMapMultiplePlayers(maximizerID, otherPlayerIDs, status, stopChannel))
}

// MinimaxClient is a client implementation that uses Minimax to decide what to do next
type MinimaxClient struct{}

// GetAction implements the Client interface
func (c MinimaxClient) GetAction(status *Status, calculationTime time.Duration) Action {
	stopChannel := time.After((calculationTime / 10) * 9)
	otherPlayerID, err := status.FindClosestPlayerTo(status.You)
	if err != nil {
		log.Println("could not find closest player:", err)
		return ChangeNothing
	}
	log.Println("using player", otherPlayerID, "at", status.Players[otherPlayerID].X, status.Players[otherPlayerID].Y, "as minimizer")
	actions := MinimaxBestActions(status.You, []uint8{otherPlayerID}, status, stopChannel)
	if len(actions) > 0 {
		return actions[rand.Intn(len(actions))]
	}
	log.Println("no best action, using change_nothing")
	return ChangeNothing
}
