package main

import (
	"log"
	"runtime"
	"time"
)

// Result :
type Result struct {
	Visits [][]uint16
	Player []*SimPlayer
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

// Simulate games for all given Players and the given SimDepth. Uses simulatePlayer and resultsToField to achieve this
func calculateProbabilityTables(status *Status, stopSimulateGameChan <-chan time.Time, activePlayersInRange []*Player) [][][]float64 {
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
	probabilityTableChannels := make(map[int]chan [][]float64, 0)
	//with those Channels simulatePlayer gives a result back to simulateGame to process the result
	resultChannels := make(map[int]chan *Result, 0)
	for j, simPlayer := range allSimPlayer {
		probabilityTableChannels[j] = make(chan [][]float64, 1)
		resultChannels[j] = make(chan *Result, maxSimDepth)
		go calculateVisitsForPlayer(simPlayer, j, status, status.Turn, resultChannels[j], probabilityTableChannels[j], stopSimulateGameChan)
	}
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
			player = nil
		}
	}
	runtime.GC()
	for o, probabilityTable := range playerProbabilityTables {
		fieldChannels[o] <- probabilityTable
	}
	return playerProbabilityTables[me]
}

// simulate all possible moves for a given simPlayer and a given status until numberOfTurns is reached
func calculateVisitsForPlayer(simPlayer *SimPlayer, id int, status *Status, elapsedTurns uint16, resultChannel chan<- *Result, probabilityTableChannel <-chan [][]float64, stopChannel <-chan time.Time) {
	var currentProbabilityTable [][]float64
	currentPlayers := make([]*SimPlayer, 1)
	currentPlayers[0] = simPlayer
	for turn := 1; turn <= maxSimDepth; turn++ {
		visits := make([][]uint16, status.Height)
		for r := range visits {
			visits[r] = make([]uint16, status.Width)
		}
		if turn != 1 {
			select {
			case newProbabilityTable := <-probabilityTableChannel:
				addFields(&currentProbabilityTable, newProbabilityTable)
			case <-stopChannel:
				log.Println("ended simulation for player", id)
				currentPlayers = nil
				resultChannel <- nil
				return
			}

		} else {
			currentProbabilityTable = makeProbabilityTable(status.Height, status.Width)
		}
		children := make([]*SimPlayer, len(currentPlayers)*5)
		childCounter := 0
		for playerCounter, player := range currentPlayers {
			select {
			case <-stopChannel:
				log.Println("ended Simulation for player", id)
				currentPlayers = nil
				children = nil
				runtime.GC()
				resultChannel <- nil
				return
			default:
				possibleActions := player.player.PossibleMoves(status.Cells, elapsedTurns+uint16(turn), player.AllVisitedCells, false)
				for _, action := range possibleActions {
					child, score := simulateMove(visits, player, action, elapsedTurns+uint16(turn), currentProbabilityTable)
					child.Probability *= 1.0/float64(len(possibleActions)) - (score / float64(len(child.LastMoveVisitedCells)))
					if child.Probability < 0 {
						continue
					}
					children[childCounter] = child
					childCounter++
				}
				//set unused parts nil
				//player.AllVisitedCells = nil
				if len(currentPlayers)-1 > 0 {
					currentPlayers = currentPlayers[1:len(currentPlayers)]
				}
				if playerCounter%processedPlayersTillGC == 0 && playerCounter != 0 && turn >= 6 {
					go runtime.GC()
				}
			}

		}
		currentPlayers = children[0:childCounter]
		children = nil
		if resultChannel != nil {
			resultChannel <- &Result{Visits: visits, Player: currentPlayers}
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
