package main

import (
	"log"
	"math"
	"sort"
	"time"
)

var probabilityTableOfLastTurn [][]float64

//this defines the window size from where the player reads the probabilites at the beginnig to analyze the field an knows if he should use minimax
const windowSize = 8

//If the sum of all probabilities in the specified window is higher then this, minimax can be used
var minimaxActivationValue = 0.008

//if minimax can be used a player also has to be nearer than this value to the player so it gets minimaxed
const minimaxDistance = 14.0

//This is the minimal Number of other players we are using to calculate the probability tables (if so many players are living)
const minimalNumberOfSimPlayers = 2

//If a player is nearer then this distance it will always be used for the calculation of the probability tables
const simPlayerDistance = 20.0

//This functions executes a action and returns the average probability of every visited Cell
func evaluateAction(player *Player, field [][]float64, action Action, turn uint16) float64 {
	probability := 0.0
	visitedCoords := player.ProcessAction(action, turn)
	counterVisitedCoords := 0
	for _, coords := range visitedCoords {
		if coords == nil {
			continue
		}
		probability += field[coords.Y][coords.X]
		counterVisitedCoords++

	}
	if counterVisitedCoords != 0 {
		return probability / float64(counterVisitedCoords)
	}
	return probability
}

//computes a score for every possible Action. The action with the lowest score is chosen
func evaluatePaths(player Player, allFields [][][]float64, paths [][]Action, turn uint16, simDepth int, possibleActions []Action, minimaxIsUsed bool) Action {
	var probabilities [5]float64
	var possible [5]bool
	//Computes if a action is possible based on the possibleActions Array
	for _, action := range possibleActions {
		possible[action] = true
	}
	var inPaths [5]bool
	//computes the score for every path
	for _, path := range paths {
		score := 0.0
		minPlayer := player.Copy()
		for i := 0; i < len(path); i++ {
			if i != len(path) {
				if i <= simDepth {
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
		inPaths[path[0]] = true
		probabilities[path[0]] += score
	}
	log.Println(possible)
	if !minimaxIsUsed {
		for z := range possible {
			possible[z] = possible[z] && inPaths[z]
		}
	}
	log.Println(possible)
	//computes how many times a Action was the first Action of path
	counter := [5]int{1, 1, 1, 1, 1}
	for _, path := range paths {
		counter[path[0]]++
	}
	var sumOfProbabilities float64
	for z, probability := range probabilities {
		sumOfProbabilities += probability / float64(counter[z])
	}
	log.Println(sumOfProbabilities)
	var scores [5]float64
	for i := 0; i < 5; i++ {
		if sumOfProbabilities != 0 {
			log.Println(probabilities[i] / float64(counter[i]) / sumOfProbabilities)
			log.Println(1.0 - (float64(counter[i]) / float64(len(paths))))
			scores[i] = (probabilities[i] / float64(counter[i]) / sumOfProbabilities) + (1.0 - (float64(counter[i]) / float64(len(paths))))
		} else {
			scores[i] = 1.0 - (float64(counter[i]) / float64(len(paths)))
		}
	}
	//computes Value based on the score of a Action an
	log.Printf("calculated values %1.2f", scores)

	minimum := math.Inf(0)
	action := ChangeNothing
	for i, v := range scores {
		if possible[i] && v < minimum {
			minimum = v
			action = Action(i)
		}
	}

	promissingPaths = make([][]Action, 0)
	for _, path := range paths {
		if path[0] == action && len(path) > 1 {
			promissingPaths = append(promissingPaths, path[1:len(path)-1])
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
		var occupiedCellsInWindow float64
		var accumulatedProbability float64
		for y := int(me.Y) - windowSize; y < int(me.Y)+windowSize; y++ {
			if y >= 0 && y < int(status.Height) {
				for x := int(me.X) - windowSize; x < int(me.X)+windowSize; x++ {
					if x >= 0 && x < int(status.Width) {
						accumulatedProbability += probabilityTable[y][x]
						if status.Cells[y][x] {
							occupiedCellsInWindow++
						} else if probabilityTable[y][x] == 0 {
							occupiedCellsInWindow++
						}
					} else {
						occupiedCellsInWindow++
					}
				}

			} else {
				occupiedCellsInWindow += 2*windowSize + 1
			}
		}
		if occupiedCellsInWindow != 0 {
			accumulatedProbability /= math.Pow((2*windowSize+1), 2.0) - occupiedCellsInWindow
		}
		log.Printf("The average probability in the window is %1.2e", accumulatedProbability)
		if accumulatedProbability >= minimaxActivationValue {
			playersAreNear = true
			simulateOtherPlayers = true
		}
	} else {
		playersAreNear = true
	}
	relativeDistanceTo := make(map[float64]*Player)
	for z, player := range status.Players {
		if player == me {
			continue
		}
		distance := player.DistanceTo(me)
		relativeDistance := distance / float64(player.Speed) / float64(me.Speed)
		if relativeDistance < minimaxDistance && playersAreNear {
			minimaxPlayers = append(minimaxPlayers, z)
		}
		relativeDistanceTo[distance/float64(player.Speed)] = player
	}
	relativeDistances := make([]float64, len(relativeDistanceTo))
	i := 0
	for relativeDistance := range relativeDistanceTo {
		relativeDistances[i] = relativeDistance
		i++
	}
	sort.Float64s(relativeDistances)
	counter := 0
	for _, distance := range relativeDistances {
		if counter <= minimalNumberOfSimPlayers || distance < simPlayerDistance {
			probabilityPlayers = append(probabilityPlayers, relativeDistanceTo[distance])
		}
	}
	probabilityPlayers = append(probabilityPlayers, status.Players[status.You])
	return minimaxPlayers, probabilityPlayers
}

// CombiClient is a client implementation that uses a combination of probability Tables, rollouts and minimax to decide what to do next
type CombiClient struct{}

// GetAction implements the Client interface
func (c CombiClient) GetAction(player Player, status *Status, calculationTime time.Duration) Action {
	start := time.Now()
	stopChannel1 := time.After(calculationTime / 10 * 6)
	stopChannel2 := time.After(calculationTime / 10 * 8)
	var bestAction Action
	possibleActions := player.PossibleActions(status.Cells, status.Turn, nil, false)
	//handle trivial cases (zero or one possible Action)
	if len(possibleActions) == 1 {
		log.Println("only possible action: ", possibleActions[0])
		return possibleActions[0]
	} else if len(possibleActions) == 0 {
		log.Println("going to die... choosing change_nothing as last action")
		return ChangeNothing
	}
	minimaxPlayers, probabilityPlayers := analyzeBoard(status, probabilityTableOfLastTurn)
	stopRolloutChan := make(chan time.Time)
	rolloutChan := make(chan [][]Action, 1)
	go func() {
		rolloutPaths := simulateRollouts(status, stopRolloutChan)
		rolloutChan <- rolloutPaths
	}()

	//calculate which players are simulated TODO: Move this code to an external function and improve it
	log.Println("using", len(probabilityPlayers), "players, for calculation of probabilityFields")
	minimaxChannel := make(chan []Action, 1)
	stopMinimaxChannel := make(chan time.Time)
	//If there is more than one player we should calculate minimax for we need minimax for mutliple players
	if len(minimaxPlayers) > 1 {
		go func() {
			log.Println("using minimax wih", len(minimaxPlayers), "opponents")
			bestActionsMinimax := minimaxBestActionsMultiplePlayers(minimaxPlayers, status.You, status, stopMinimaxChannel)
			minimaxChannel <- bestActionsMinimax
		}()
	} else if len(minimaxPlayers) == 1 {
		go func() {
			log.Println("using minimax with one opponent")
			log.Println("using player", minimaxPlayers[0], "at", status.Players[minimaxPlayers[0]].X, status.Players[minimaxPlayers[0]].Y, "as minimizer")
			bestActionsMinimax := MinimaxBestActionsTimed(status.You, minimaxPlayers[0], status, stopMinimaxChannel)
			minimaxChannel <- bestActionsMinimax
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
	_ = <-stopChannel1
	log.Println("sending stop signal to simulateGame...")
	close(stopCalculateProbabilityTables)
	_ = <-stopChannel2
	log.Println("sending stop signal to simulateRollouts and minimax...")
	close(stopRolloutChan)
	close(stopMinimaxChannel)
	var useMinimax bool
	if len(minimaxPlayers) > 0 {
		minimaxActions := <-minimaxChannel
		//If we use minimax against multiple players minimax actions might be empty. Then we use possible actions
		if len(minimaxActions) != 0 {
			possibleActions = minimaxActions
			useMinimax = true
		}
	}
	allProbabilityTables = <-probabilityTablesChan
	bestPaths := <-rolloutChan

	log.Println("found", len(bestPaths), "paths that should be evaluated")
	log.Println("could calculate probabilityTables for", len(allProbabilityTables), "turns")
	//This is only for debugging purposes and combines the last field with the status
	//log.Println(allProbabilityTables[len(allProbabilityTables)-1])
	//log.Println("Last calculated probability Table")
	//for y, row := range allProbabilityTables[len(allProbabilityTables)-1] {
	//fmt.Printf("%2d, %1.1e\n", y, row)
	//}
	//Log Timing
	log.Println("time until calculations are finished and evaluation can start: ", time.Since(start))
	//Evaluate the paths with the given field and return the best Action based on this TODO: Needs improvement in case of naming
	bestAction = evaluatePaths(player, allProbabilityTables, bestPaths, status.Turn, len(allProbabilityTables)-1, possibleActions, useMinimax)
	//Log Timing
	probabilityTableOfLastTurn = allProbabilityTables[len(allProbabilityTables)-1]
	totalProcessingTime := time.Since(start)
	simulateOtherPlayers = false
	if totalProcessingTime > calculationTime {
		panic("Couldn't reach timing goal")
	}
	log.Println("total processing took", totalProcessingTime)
	log.Println("chose best action", bestAction)
	return bestAction
}
