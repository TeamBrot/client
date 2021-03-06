package main

import (
	"log"
	"math"
	"runtime"
	"time"
)

// This const defines the startProbability for every player besides us. There is no good reason to change it
const othersStartProbability = 1.0

//This const defines the maximal number of Turns simulateGame will try to process
const maxSimDepth = 40

//This const defines after how many processed players simulatePlayer will schedule a garbage Collection cycle. Lowering the value improves memory efficiency but has a performance impact
const processedPlayersTillGC = 60000

// Result :
type Result struct {
	Visits [][]uint16
	Player []*SimPlayer
}

//Simulates a Action and for a simPlayer and a visitTable. Raises the score of every visited cell at the visitTable and adds the Coords to allVisitedCells and lastMoveVisitedCells of the simPlayer
func simulateAction(visitTable [][]uint16, parentPlayer *SimPlayer, action Action, turn uint16, simField [][]float64) (*SimPlayer, float64) {
	childPlayer := parentPlayer.Copy()
	player := childPlayer.player
	score := 0.0
	visitedCoords := player.ProcessAction(action, turn)
	for _, coord := range visitedCoords {
		if coord == nil {
			continue
		}
		visitTable[coord.Y][coord.X]++
		childPlayer.AllVisitedCells[*coord] = struct{}{}
		childPlayer.LastMoveVisitedCells[*coord] = struct{}{}
		score += simField[coord.Y][coord.X]
	}
	return childPlayer, score
}

// calculates the probabilityTables for all given Players. It coordinates the timing between calculateVisitsForPlayer and visitsToProbabilities. After every iteration of calculateVisitsForPlayer visitsToProbabilities is called
func calculateProbabilityTables(status *Status, stopSimulateGameChan <-chan time.Time, activePlayersInRange []*Player, myStartProbability float64) [][][]float64 {
	var me int
	allSimPlayer := make([]*SimPlayer, 0)
	//Convert all Players that should be simulated to simPlayers
	for d, player := range activePlayersInRange {
		if player == status.Players[status.You] {
			// Store our index and use the specified startProbability for us instead of the startProbability for other players
			me = d
			allSimPlayer = append(allSimPlayer, SimPlayerFromPlayer(player, myStartProbability))
		} else {
			allSimPlayer = append(allSimPlayer, SimPlayerFromPlayer(player, othersStartProbability))
		}
	}
	//with those Channels visitsToProbabilities can give probabilityTables to calculateVisitsForPlayer to make the calculations more acurate
	probabilityTableChannels := make(map[int]chan [][]float64, 0)
	//with those Channels calculateVisitsForPlayer gives a result back to calculateProbabilityTables to process the result
	resultChannels := make(map[int]chan *Result, 0)
	for j, simPlayer := range allSimPlayer {
		probabilityTableChannels[j] = make(chan [][]float64, 1)
		resultChannels[j] = make(chan *Result, maxSimDepth)
		go calculateVisitsForPlayer(simPlayer, j, status, status.Turn, resultChannels[j], probabilityTableChannels[j], stopSimulateGameChan)
	}

	// After we initialized calculateVisitsForPlayer we wait for the results of every Player for every round so we can start visitsToProbabilities
	allProbabilityTables := make([][][]float64, 0)
	for z := 0; z < maxSimDepth; z++ {
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
			log.Println("The last recieved fields weren't valid")
			break
		}
		probabilityTable := visitsToProbabilities(me, results, status.Width, status.Height, probabilityTableChannels)
		allProbabilityTables = append(allProbabilityTables, probabilityTable)
		if z != 0 {
			addProbabilityTables(&allProbabilityTables[z], allProbabilityTables[z-1])
		}
	}
	return allProbabilityTables
}

//After every turn the given results are evaluated and fields are computed on basis of them
func visitsToProbabilities(me int, results []*Result, width uint16, height uint16, probabilityTablesChannel map[int]chan [][]float64) [][]float64 {
	if len(results) == 0 {
		log.Println(results)
		panic("Can't calculate probability Table without results")
	}

	//Prepares the accumulated visits array
	accumulatedVisits := make([][][]uint16, len(results))
	for u := 0; u < len(results); u++ {
		accumulatedVisits[u] = makeEmptyVisits(height, width)
	}

	//Calculates for every player, the visits in absolute numbers, that another player could have visited a cell
	for m, cells := range accumulatedVisits {
		for n, result := range results {
			if n != m {
				addVisits(&cells, result.Visits)
			}
		}
	}

	//Adjust the probability if a player tries to visit a field, that another player also could visit
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

	//After the probabilites have been adjusted we write the probabilities for every player into the probability table
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
		}
	}

	//Give back the probability for the visit calculation
	for o, probabilityTable := range playerProbabilityTables {
		probabilityTablesChannel[o] <- probabilityTable
	}

	return playerProbabilityTables[me]
}

// simulate all possible actions for a given simPlayer and a given status until numberOfTurns is reached
func calculateVisitsForPlayer(simPlayer *SimPlayer, id int, status *Status, elapsedTurns uint16, resultChannel chan<- *Result, probabilityTableChannel <-chan [][]float64, stopChannel <-chan time.Time) {
	var currentProbabilityTable [][]float64
	currentPlayers := make([]*SimPlayer, 1)

	currentPlayers[0] = simPlayer
	for turn := 0; turn < maxSimDepth; turn++ {
		//Initialize the visit table for this turn
		visitTable := make([][]uint16, status.Height)
		for r := range visitTable {
			visitTable[r] = make([]uint16, status.Width)
		}
		if turn != 0 {
			//recieve a new probabilityTable for this player from visits to probabilities or break if recieving the stop signal
			select {
			case newProbabilityTable := <-probabilityTableChannel:
				addProbabilityTables(&currentProbabilityTable, newProbabilityTable)
			case <-stopChannel:
				currentPlayers = nil
				resultChannel <- nil
				log.Println("ended simulation for player", id)
				return
			}
		} else {
			//Initialize the first probability Table
			currentProbabilityTable = makeProbabilityTable(status.Height, status.Width)
		}
		children := make([]*SimPlayer, len(currentPlayers)*5)
		childCounter := 0
		//Finding the child for every current player
		for playerCounter, player := range currentPlayers {
			select {
			//stopping if we recieve the stop signal
			case <-stopChannel:
				currentPlayers = nil
				children = nil
				resultChannel <- nil
				log.Println("ended Simulation for player", id)
				return
			default:
				//Make a child for every possible Action
				possibleActions := player.player.PossibleActions(status.Cells, elapsedTurns+uint16(turn), player.AllVisitedCells, false)
				for _, action := range possibleActions {
					child, score := simulateAction(visitTable, player, action, elapsedTurns+uint16(turn), currentProbabilityTable)
					child.Probability *= 1.0/float64(len(possibleActions)) - (score / float64(len(child.LastMoveVisitedCells)))
					//the child probability could get zero or below, if other player also try to visit the fields this child tries to use
					if child.Probability <= 0 {
						continue
					}
					children[childCounter] = child
					childCounter++
				}
				//Set the now unused part of the currentPlayer array to nil
				if len(currentPlayers)-1 > 0 {
					currentPlayers = currentPlayers[1:len(currentPlayers)]
				}
				//schedule the garbage collection if the condition is met. This costs performance but improves memory usage by a lot
				if playerCounter%processedPlayersTillGC == 0 && playerCounter != 0 && turn >= 6 {
					go runtime.GC()
				}
			}

		}
		//Set new currentPlayers
		currentPlayers = children[0:childCounter]
		children = nil
		//send back the result of the calculation to calculate probabilityTables
		if resultChannel != nil {
			resultChannel <- &Result{Visits: visitTable, Player: currentPlayers}
		}
		runtime.GC()
	}

	log.Println("finished calculation for player", id)
	close(resultChannel)
	return
}

//Adds second field to first Field. First field has to be a pointer and is going to be changed
func addProbabilityTables(field1 *[][]float64, field2 [][]float64) {
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

// returns an empty visitTable
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

//ProbabilityClient is a client implementation that uses a probabilityTable to decide what to do next
type ProbabilityClient struct {
	myStartProbability float64
}

// GetAction implements the Client interface
func (c ProbabilityClient) GetAction(status *Status, calculationTime time.Duration) Action {
	stopChannel := time.After((calculationTime / 10) * 7)
	possibleActions := status.Players[status.You].PossibleActions(status.Cells, status.Turn, nil, false)
	if len(possibleActions) == 0 {
		log.Println("I'll die")
		return ChangeNothing
	}
	var allPlayers []*Player
	for _, player := range status.Players {
		allPlayers = append(allPlayers, player)
	}
	allProbabilityTables := calculateProbabilityTables(status, stopChannel, allPlayers, c.myStartProbability)
	var possible [5]bool
	//Computes if a action is possible based on the possibleActions Array
	for _, action := range possibleActions {
		possible[action] = true
	}
	var values [5]float64
	for _, action := range possibleActions {
		values[action] = evaluateAction(status.Players[status.You].Copy(), allProbabilityTables[0], action, status.Turn)

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
