package main

import (
	"log"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"time"
)

// Result :
type Result struct {
	Visits [][]uint16
	Player []*SimPlayer
}

//Coords store the coordinates of a player
type Coords struct {
	Y, X uint16
}

//SimPlayer to add a new array of visited cells
type SimPlayer struct {
	player               *Player
	Probability          float64
	AllVisitedCells      map[Coords]struct{}
	LastMoveVisitedCells map[Coords]struct{}
}

//Simulates a Action and for a simPlayer and a board. Raises the score of every visited cell at the board and adds the Coords to allVisitedCells and lastMoveVisitedCells
func simulateMove(board [][]uint16, parentPlayer *SimPlayer, action Action, turn uint16, simField [][]float64) (*SimPlayer, float64) {
	childPlayer := parentPlayer.copySimPlayer()
	player := childPlayer.player
	score := 0.0
	visitedCoords := player.ProcessAction(action, turn)
	for _, coord := range visitedCoords {
		if coord == nil {
			continue
		}
		board[coord.Y][coord.X]++
		childPlayer.AllVisitedCells[*coord] = struct{}{}
		childPlayer.LastMoveVisitedCells[*coord] = struct{}{}
		score += simField[coord.Y][coord.X]
	}
	return childPlayer, score
}

//implements the doMove function for the rollout function
func rolloutMove(status *Status, action Action, player *Player) {
	visitedCoords := player.ProcessAction(action, status.Turn)
	for _, coords := range visitedCoords {
		if coords == nil {
			continue
		}
		status.Cells[coords.Y][coords.X] = true
	}

}

//search for the longest paths a player could reach. Simulates random move for all Players and allways processes as last player
func simulateRollouts(status *Status, limit int, filterValue float64, stopSimulateRollouts <-chan time.Time) [][]Action {
	longest := 0
	maxNumberofRollouts := 7000000
	longestPaths := make([][]Action, 0, maxNumberofRollouts)
	for j := 0; j < maxNumberofRollouts; j++ {
		select {
		case <-stopSimulateRollouts:
			log.Println("could perfom", j, "rollouts")
			log.Println("The longest path was", longest, "Actions long")
			return longestPaths
		default:
			rolloutStatus := status.Copy()
			path := make([]Action, 0)
			counter := 0
			for {
				me := rolloutStatus.Players[status.You]
				//countLivingPlayers := 0
				rolloutStatus.Turn++
				possibleMoves := me.PossibleMoves(rolloutStatus.Cells, rolloutStatus.Turn, nil, false)
				if len(possibleMoves) == 0 {
					break
				}
				var randomAction Action
				//This should distribute the first Action taken equally
				if counter == 0 {
					randomAction = possibleMoves[j%len(possibleMoves)]
					counter++
				} else {
					randomAction = possibleMoves[rand.Intn(len(possibleMoves))]
				}
				rolloutMove(rolloutStatus, randomAction, me)
				path = append(path, randomAction)
			}
			//Now we chek if the last taken path was longer then every other path
			if float64(len(path)) >= float64(longest)*filterValue {
				//if longest is still bigger then the path found now we only append the path
				if longest >= len(path) {
					longestPaths = append(longestPaths, path)
					//If it is bigger by a lot we can forget every path we found until now
				} else if float64(len(path))*filterValue > float64(longest) {
					longestPaths = make([][]Action, 0, maxNumberofRollouts-j)
					longestPaths = append(longestPaths, path)
					longest = len(path)
					//If none of the before is the case we have to filter all values that are in longest paths until now
				} else {
					longestPaths = filterPaths(longestPaths, len(path), filterValue, maxNumberofRollouts-j+len(longestPaths))
					longestPaths = append(longestPaths, path)
					longest = len(path)
				}
			}
		}
	}
	log.Println("could perfom", maxNumberofRollouts, "rollouts, which is the maximum possible")
	return longestPaths
}

//Filters an given array of paths and returns an array of paths that match the criteria
func filterPaths(paths [][]Action, longest int, percent float64, maxRemaining int) [][]Action {
	filteredPaths := make([][]Action, 0, maxRemaining)
	for _, path := range paths {
		if float64(len(path)) >= float64(longest)*percent {
			filteredPaths = append(filteredPaths, path)
		}
	}
	return filteredPaths
}

// Copys a SimPlayer Object (Might be transfered to a util.go)
func (player *SimPlayer) copySimPlayer() *SimPlayer {
	var p SimPlayer
	p.player = player.player.copyPlayer()
	p.Probability = player.Probability
	p.LastMoveVisitedCells = make(map[Coords]struct{})
	p.AllVisitedCells = make(map[Coords]struct{})
	for k := range player.AllVisitedCells {
		p.AllVisitedCells[k] = struct{}{}
	}
	return &p
}

// Simulate games for all given Players and the given SimDepth. Uses simulatePlayer and resultsToField to achieve this
func simulateGame(status *Status, stopSimulateGameChan <-chan time.Time, simDepth int, activePlayersInRange []*Player) [][][]float64 {
	var me int
	allSimPlayer := make([]*SimPlayer, 0)
	//Convert all Players that should be simulated to simPlayers
	for d, player := range activePlayersInRange {
		if player == status.Players[status.You] {
			me = d
			allSimPlayer = append(allSimPlayer, player.toSimPlayer(1.2))
		} else {
			allSimPlayer = append(allSimPlayer, player.toSimPlayer(1.0))
		}
	}
	//with those Channels resultsToField can give fields to simulatePlayer to make the calculations more acurate
	visitedCellsChannels := make(map[int]chan [][]float64, 0)
	//with those Channels simulatePlayer gives a result back to simulateGame to process the result
	resultChannels := make(map[int]chan *Result, 0)
	for j, simPlayer := range allSimPlayer {
		visitedCellsChannels[j] = make(chan [][]float64, 1)
		resultChannels[j] = make(chan *Result, simDepth)
		go simulatePlayer(simPlayer, j, status, simDepth, status.Turn, resultChannels[j], visitedCellsChannels[j], stopSimulateGameChan)
	}
	allProbabilityTables := make([][][]float64, 0)
	for z := 0; z < simDepth; z++ {
		results := make([]*Result, len(resultChannels))
		valid := true
		for l, ch := range resultChannels {
			select {
			case <-stopSimulateGameChan:
				log.Println("ended simulateGame, returning field")
				return allProbabilityTables
			case val := <-ch:
				if val != nil {

					results[l] = val
				} else {
					valid = false
					break
				}
			}
		}
		if !valid {
			break
		}
		log.Println("starting to calculate field for turn", z)
		probabilityTable := visitsToProbabilities(me, results, status.Width, status.Height, visitedCellsChannels)
		allProbabilityTables = append(allProbabilityTables, probabilityTable)
		if z != 0 {
			addFields(&allProbabilityTables[z], allProbabilityTables[z-1])
		}
	}
	return allProbabilityTables
}

//After every turn the given results are evaluated and fields are computed on basis of them
func visitsToProbabilities(me int, results []*Result, width uint16, height uint16, fieldChannels map[int]chan [][]float64) [][]float64 {
	if len(results) == 0 {
		log.Println(results)
		panic("Can't calculate probability Table without results")
	}
	accumulatedVisits := make([][][]uint16, len(results))
	for u := 0; u < len(results); u++ {
		accumulatedVisits[u] = makeEmptyVisits(height, width)
	}

	for m, cells := range accumulatedVisits {
		for n, result := range results {
			if n != m {
				addVisits(&cells, result.Visits)
			}
		}
	}

	for g, cells := range accumulatedVisits {
		result := results[g]
		for _, player := range result.Player {
			if player == nil {
				break
			}
			for coord := range player.LastMoveVisitedCells {
				if cells[coord.Y][coord.X] != 0 {
					player.Probability /= float64(cells[coord.Y][coord.X])
				}
			}
		}
	}

	playerProbabilityTables := make([][][]float64, len(results))
	for k := range playerProbabilityTables {
		playerProbabilityTables[k] = makeProbabilityTable(height, width)
	}

	for l, result := range results {
		for _, player := range result.Player {
			if player == nil {
				break
			}
			for coords := range player.LastMoveVisitedCells {
				for z, field := range playerProbabilityTables {
					if z != l {
						field[coords.Y][coords.X] += player.Probability
					}
				}
			}
			player.LastMoveVisitedCells = nil
			player = nil
		}
	}

	for o, probabilityTable := range playerProbabilityTables {
		fieldChannels[o] <- probabilityTable
	}
	runtime.GC()
	return playerProbabilityTables[me]

}

// simulate all possible moves for a given simPlayer and a given status until numberOfTurns is reached
func simulatePlayer(simPlayer *SimPlayer, id int, status *Status, numberOfTurns int, elapsedTurns uint16, resultChannel chan<- *Result, boardChannel <-chan [][]float64, stopChannel <-chan time.Time) {
	var fieldAfterTurn [][]float64
	currentPlayers := make([]*SimPlayer, 1)
	currentPlayers[0] = simPlayer
	for i := 0; i < numberOfTurns; i++ {
		turn := i + 1
		writeField := make([][]uint16, status.Height)
		for r := range writeField {
			writeField[r] = make([]uint16, status.Width)
		}
		if i != 0 {
			select {
			case newField := <-boardChannel:
				addFields(&fieldAfterTurn, newField)
			case <-stopChannel:
				currentPlayers = nil
				resultChannel <- nil
				return
			}

		} else {
			fieldAfterTurn = makeProbabilityTable(status.Height, status.Width)
		}
		children := make([]*SimPlayer, len(currentPlayers)*5)
		counter := 0
		for _, player := range currentPlayers {
			select {
			case <-stopChannel:
				log.Println("ended Simulation for player", id)
				currentPlayers = nil
				children = nil
				resultChannel <- nil
				return
			default:
				possibleActions := player.player.PossibleMoves(status.Cells, elapsedTurns+uint16(turn), player.AllVisitedCells, false)
				for _, action := range possibleActions {
					child, score := simulateMove(writeField, player, action, elapsedTurns+uint16(turn), fieldAfterTurn)
					child.Probability *= 1.0/float64(len(possibleActions)) - (score / float64(len(child.LastMoveVisitedCells)))
					if child.Probability < 0 {
						continue
					}
					children[counter] = child
					counter++
				}
				player.AllVisitedCells = nil
				player.player = nil

			}

		}
		currentPlayers = children[0:counter]
		children = nil
		if resultChannel != nil {
			resultChannel <- &Result{Visits: writeField, Player: currentPlayers}
		}
		runtime.GC()
	}
	log.Println("finished calculation for player", id)
	close(resultChannel)
	return
}

//Adds second field to first Field. First field has to be a pointer and is going to be changed
func addFields(field1 *[][]float64, field2 [][]float64) {
	field := *field1
	for i := 0; i < len(field); i++ {
		for j := 0; j < len(field[i]); j++ {
			field[i][j] += field2[i][j]
		}
	}

}

//Adds second visit table to first visit table. First board has to be a pointer an is going to be changed!
func addVisits(field1 *[][]uint16, field2 [][]uint16) {
	field := *field1
	for i := 0; i < len(field); i++ {
		for j := 0; j < len(field[i]); j++ {
			field[i][j] += field2[i][j]
		}
	}

}

// returns an empty field
func makeEmptyVisits(height uint16, width uint16) [][]uint16 {
	field := make([][]uint16, height)
	for r := range field {
		field[r] = make([]uint16, width)
	}
	return field
}

func makeProbabilityTable(height uint16, width uint16) [][]float64 {
	field := make([][]float64, height)
	for r := range field {
		field[r] = make([]float64, width)
	}
	return field
}

//This functions executes a action and returns the average score of every visited Cell
func evaluateAction(player *Player, field [][]float64, action Action, turn uint16) float64 {
	score := 0.0
	visitedCoords := player.ProcessAction(action, turn)
	for _, coords := range visitedCoords {
		if coords == nil {
			continue
		}
		score += field[coords.Y][coords.X]

	}
	return score / float64(player.Speed)
}

//This function copies a struct of type Player
func (oldPlayer *Player) copyPlayer() *Player {
	var newPlayer Player
	newPlayer.Direction = oldPlayer.Direction
	newPlayer.Speed = oldPlayer.Speed
	newPlayer.X = oldPlayer.X
	newPlayer.Y = oldPlayer.Y
	return &newPlayer
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
func analyzeBoard(status *Status) ([]uint8, []*Player) {
	var probabilityPlayers []*Player
	var minimaxPlayers []uint8
	//var turn uint16
	//turn = 0
	me := status.Players[status.You]
	if len(status.Players) <= 2 {
		for i, player := range status.Players {
			probabilityPlayers = append(probabilityPlayers, player)
			if player != me {
				minimaxPlayers = append(minimaxPlayers, i)
			}
		}
		return minimaxPlayers, probabilityPlayers
	}
	distanceTo := make(map[float64]*Player)
	for z, player := range status.Players {
		if player == me {
			continue
		}
		distance := player.DistanceTo(me)
		relativeDistance := distance / float64(player.Speed) / float64(me.Speed)
		if relativeDistance < 12.0 {
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
		if counter < 3 || distance < 20 {
			probabilityPlayers = append(probabilityPlayers, distanceTo[distance])
		}
		if distance < 12 {

			minimaxPlayers = append(minimaxPlayers)
		}
	}
	probabilityPlayers = append(probabilityPlayers, status.Players[status.You])
	return minimaxPlayers, probabilityPlayers
}

func spekuTiming(calculationTime time.Duration, timingChannel chan<- time.Time) {
	time.Sleep(time.Duration(0.4 * float64(calculationTime.Nanoseconds())))
	timingChannel <- time.Now()
	time.Sleep(time.Duration(0.4 * float64(calculationTime.Nanoseconds())))
	close(timingChannel)
}

func (player *Player) toSimPlayer(probability float64) *SimPlayer {
	var simPlayer SimPlayer
	simPlayer.player = player.copyPlayer()
	simPlayer.AllVisitedCells = make(map[Coords]struct{}, 0)
	simPlayer.LastMoveVisitedCells = make(map[Coords]struct{}, 0)
	simPlayer.Probability = probability
	return &simPlayer
}

// SpekuClient is a client implementation that uses speculation to decide what to do next
type SpekuClient struct{}

// GetAction implements the Client interface
func (c SpekuClient) GetAction(player Player, status *Status, calculationTime time.Duration) Action {
	start := time.Now()
	timingChannel := make(chan time.Time)

	go spekuTiming(calculationTime, timingChannel)
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
		rolloutPaths := simulateRollouts(status, 300, 0.75, stopRolloutChan)
		rolloutChan <- rolloutPaths
	}()

	//calculate which players are simulated TODO: Move this code to an external function and improve it
	minMaxPlayers, probabilityPlayers := analyzeBoard(status)
	log.Println("simulating", len(probabilityPlayers), "players")

	miniMaxChannel := make(chan []Action, 1)
	stopMiniMaxChannel := make(chan time.Time)
	if len(minMaxPlayers) > 1 {
		go func() {
			bestActionsMinimax := miniMaxMultiplePlayers(minMaxPlayers, status.You, status, stopMiniMaxChannel)
			miniMaxChannel <- bestActionsMinimax
		}()
	} else if len(minMaxPlayers) == 1 {
		go func() {
			bestActionsMinimax := MinimaxBestActionsTimed(status.You, minMaxPlayers[0], status, stopMiniMaxChannel)
			miniMaxChannel <- bestActionsMinimax
		}()
	}
	var maxSimDepth int
	//if Your Computer is really beefy it might be a good idea to set this higher (else it is not and your computer will crash!!)
	maxSimDepth = 9
	//If this channel is closed, it will try to end simulate game

	//This channel is used to recieve an array of all calculated Fields from simulate game
	var allProbabilityTables [][][]float64
	var probabilityTablesChan chan [][][]float64
	var stopSimulateGameChan chan time.Time
	if len(probabilityPlayers) > 1 {
		stopSimulateGameChan = make(chan time.Time)
		probabilityTablesChan = make(chan [][][]float64, 1)
		go func() {
			probabilityTables := simulateGame(status, stopSimulateGameChan, maxSimDepth, probabilityPlayers)
			probabilityTablesChan <- probabilityTables
		}()
	} else {
		allProbabilityTables = make([][][]float64, maxSimDepth)
		for z := range allProbabilityTables {
			allProbabilityTables[z] = makeProbabilityTable(status.Height, status.Width)
		}
	}

	_ = <-timingChannel
	if len(probabilityPlayers) > 1 {
		log.Println("sending stop signal to simulateGame...")
		close(stopSimulateGameChan)
	}
	_ = <-timingChannel
	log.Println("sending stop signal to simulateRollouts and minimax...")
	close(stopRolloutChan)
	close(stopMiniMaxChannel)
	if len(minMaxPlayers) > 0 {
		possibleActions = <-miniMaxChannel
	} else {
		possibleActions = status.Players[status.You].PossibleMoves(status.Cells, status.Turn, nil, false)
	}
	if len(probabilityPlayers) > 1 {
		allProbabilityTables = <-probabilityTablesChan
	}
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
	log.Println("total processing took", time.Since(start))
	log.Println("chose best action", bestAction)
	return bestAction
}
