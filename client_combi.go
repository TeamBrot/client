package main

import (
	"log"
	"math"
	"sort"
	"time"
)

var probabilityTableOfLastTurn [][]float64

//Coords store the coordinates of a player
type Coords struct {
	Y, X uint16
}

//This functions executes a action and returns the average score of every visited Cell
func evaluateAction(player *Player, field [][]float64, action Action, turn uint16) float64 {
	score := 0.0
	visitedCoords := player.ProcessAction(action, turn)
	counterVisitedCoords := 0
	for _, coords := range visitedCoords {
		if coords == nil {
			continue
		}
		score += field[coords.Y][coords.X]
		counterVisitedCoords++

	}
	if counterVisitedCoords != 0 {
		return score / float64(counterVisitedCoords)
	}
	return score
}

//Simulates the moves of all Longest Paths until simDepth is reached. Computes a score for every possible Action and returns the Action with the lowes score
func evaluatePaths(player Player, allFields [][][]float64, paths [][]Action, turn uint16, simDepth int, possibleActions []Action) Action {
	var scores [5]float64
	var possible [5]bool
	//Computes if a action is possible based on the possibleActions Array
	for _, action := range possibleActions {
		possible[action] = true
	}
	turn++
	//computes the score for every path
	for _, path := range paths {
		score := 0.0
		minPlayer := player.copyPlayer()
		for i := 0; i < len(path); i++ {
			if i != len(path) {
				if i < simDepth {
					score += evaluateAction(minPlayer, allFields[i], path[i], turn+uint16(i))
				} else {
					score += evaluateAction(minPlayer, allFields[simDepth-1], path[i], turn+uint16(i))
				}
			} else {
				break
			}
		}
		if len(path) == 0 {
			log.Println("all other players are going to die in the next turn")
			return possibleActions[0]
		}
		score /= float64(len(path))
		scores[path[0]] += score
	}
	//computes how many times a Action was the first Action of path
	counter := [5]int{1, 1, 1, 1, 1}
	for _, path := range paths {
		counter[path[0]]++
	}
	var values [5]float64
	for i := 0; i < 5; i++ {
		values[i] = (scores[i] / float64(counter[i])) + (1.0 - (float64(counter[i]) / float64(len(paths))))
	}
	//computes Value based on the score of a Action an
	log.Println("calculated values", values)

	minimum := math.Inf(0)
	action := ChangeNothing
	for i, v := range values {
		if possible[i] && v < minimum {
			minimum = v
			action = Action(i)
		}
	}
	return action
}

//This Method and tells us which players we should simulate
func analyzeBoard(status *Status, probabilityTable [][]float64) ([]uint8, []*Player) {
	var probabilityPlayers []*Player
	var minimaxPlayers []uint8
	var playersAreNear bool
	me := status.Players[status.You]
	if probabilityTable != nil {
		var score float64
		for y := me.Y - 5; y < me.Y+5; y++ {
			if y >= 0 && y < status.Height {
				for x := me.X - 5; x < me.X+5; x++ {
					if x >= 0 && x < status.Width {
						score += probabilityTable[y][x]
					}
				}

			}
		}
		if score > 2 {
			playersAreNear = true
		}
	} else {
		playersAreNear = true
	}
	distanceTo := make(map[float64]*Player)
	for z, player := range status.Players {
		if player == me {
			continue
		}
		distance := player.DistanceTo(me)
		relativeDistance := distance / float64(player.Speed) / float64(me.Speed)
		if relativeDistance < 12.0 && playersAreNear {
			minimaxPlayers = append(minimaxPlayers, z)
		}
		distanceTo[distance/float64(player.Speed)] = player
	}
	distances := make([]float64, len(distanceTo))
	i := 0
	for k := range distanceTo {
		distances[i] = k
		i++
	}
	sort.Float64s(distances)
	counter := 0
	for _, distance := range distances {
		if counter < 3 || distance < 20.0 {
			probabilityPlayers = append(probabilityPlayers, distanceTo[distance])
		}
		if distance < 12.0 {
			minimaxPlayers = append(minimaxPlayers)
		}
	}
	probabilityPlayers = append(probabilityPlayers, status.Players[status.You])
	return minimaxPlayers, probabilityPlayers
}

func combiClientTiming(calculationTime time.Duration, timingChannel chan<- time.Time) {
	time.Sleep(time.Duration(0.6 * float64(calculationTime.Nanoseconds())))
	timingChannel <- time.Now()
	time.Sleep(time.Duration(0.2 * float64(calculationTime.Nanoseconds())))
	close(timingChannel)
}

// SpekuClient is a client implementation that uses speculation to decide what to do next
type SpekuClient struct{}

// GetAction implements the Client interface
func (c SpekuClient) GetAction(player Player, status *Status, calculationTime time.Duration) Action {
	start := time.Now()
	timingChannel := make(chan time.Time)
	go combiClientTiming(calculationTime, timingChannel)
	var bestAction Action
	possibleActions := player.PossibleMoves(status.Cells, status.Turn, nil, false)
	//handle trivial cases (zero or one possible Action)
	if len(possibleActions) == 1 {
		log.Println("only possible action: ", possibleActions[0])
		return possibleActions[0]
	} else if len(possibleActions) == 0 {
		log.Println("going to die... choosing change_nothing as last action")
		return ChangeNothing
	}
	stopRolloutChan := make(chan time.Time)
	rolloutChan := make(chan [][]Action, 1)
	go func() {
		rolloutPaths := simulateRollouts(status, stopRolloutChan)
		rolloutChan <- rolloutPaths
	}()

	//calculate which players are simulated TODO: Move this code to an external function and improve it
	minMaxPlayers, probabilityPlayers := analyzeBoard(status, probabilityTableOfLastTurn)
	log.Println("simulating", len(probabilityPlayers), "players")
	miniMaxChannel := make(chan []Action, 1)
	stopMiniMaxChannel := make(chan time.Time)
	//If there is more than one player we should calculate miniMax for we need minimax for mutliple players
	if len(minMaxPlayers) > 1 {
		go func() {
			bestActionsMinimax := miniMaxBestActionsMultiplePlayers(minMaxPlayers, status.You, status, stopMiniMaxChannel)
			miniMaxChannel <- bestActionsMinimax
		}()
	} else if len(minMaxPlayers) == 1 {
		go func() {
			log.Println("using player", minMaxPlayers[0], "at", status.Players[minMaxPlayers[0]].X, status.Players[minMaxPlayers[0]].Y, "as minimizer")
			bestActionsMinimax := MinimaxBestActionsTimed(status.You, minMaxPlayers[0], status, stopMiniMaxChannel)
			miniMaxChannel <- bestActionsMinimax
		}()
	}

	var allProbabilityTables [][][]float64
	//This channel is used to recieve an array of all calculated ProbabilityTables from simulate game
	var probabilityTablesChan chan [][][]float64
	var stopCalculateProbabilityTables chan time.Time
	stopCalculateProbabilityTables = make(chan time.Time)
	probabilityTablesChan = make(chan [][][]float64, 1)
	go func() {
		probabilityTables := calculateProbabilityTables(status, stopCalculateProbabilityTables, probabilityPlayers)
		probabilityTablesChan <- probabilityTables
	}()
	//recieve the first Timing signal and close the probability Calculation
	_ = <-timingChannel
	log.Println("sending stop signal to simulateGame...")
	close(stopCalculateProbabilityTables)
	_ = <-timingChannel
	log.Println("sending stop signal to simulateRollouts and minimax...")
	close(stopRolloutChan)
	close(stopMiniMaxChannel)
	if len(minMaxPlayers) > 0 {
		possibleActions = <-miniMaxChannel
	}
	allProbabilityTables = <-probabilityTablesChan
	bestPaths := <-rolloutChan

	log.Println("found", len(bestPaths), "paths that should be evaluated")
	log.Println("could calculate", len(allProbabilityTables), "turns")
	//This is only for debugging purposes and combines the last field with the status
	//log.Println(allProbabilityTables[len(allProbabilityTables)-1])
	//Log Timing
	log.Println("time until calculations are finished and evaluation can start: ", time.Since(start))
	//Evaluate the paths with the given field and return the best Action based on this TODO: Needs improvement in case of naming
	bestAction = evaluatePaths(player, allProbabilityTables, bestPaths, status.Turn, len(allProbabilityTables)-1, possibleActions)
	//Log Timing
	probabilityTableOfLastTurn = allProbabilityTables[len(allProbabilityTables)-1]
	log.Println("total processing took", time.Since(start))
	log.Println("chose best action", bestAction)
	return bestAction
}
