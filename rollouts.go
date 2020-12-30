package main

import (
	"log"
	"math/rand"
	"time"
)

//If this value is set to true we process in every rollout before we choose our own action a action for every other living player
const simulateOtherPlayers = false

//This const defines the max number of Rollouts simulateRollouts will perform. Normally there is no good reason to change this value
const maxNumberofRollouts = 7000000

//This const defines the relation between the longest and the shortest path simulateRollouts gives back
const filterValue = 0.75

//search for the longest paths a player could reach. Simulates random move for all Players and allways processes as last player
func simulateRollouts(status *Status, stopSimulateRollouts <-chan time.Time) [][]Action {
	longest := 0
	longestPaths := make([][]Action, 0, maxNumberofRollouts)
	for performedRollouts := 0; performedRollouts < maxNumberofRollouts; performedRollouts++ {
		select {
		case <-stopSimulateRollouts:
			log.Println("could perfom", performedRollouts, "rollouts")
			log.Println("The longest path was", longest, "Actions long")
			return longestPaths
		default:
			rolloutStatus := status.Copy()
			path := make([]Action, 0)
			counter := 0
			for {
				me := rolloutStatus.Players[status.You]
				rolloutStatus.Turn++
				if simulateOtherPlayers == true {
					countLivingPlayers := 0
					//Process one random move for every other player besides me
					for _, player := range rolloutStatus.Players {
						if player != me && player != nil {
							possibleMoves := player.PossibleMoves(rolloutStatus.Cells, rolloutStatus.Turn, nil, false)
							if len(possibleMoves) == 0 {
								player = nil
								continue
							}
							randomAction := possibleMoves[rand.Intn(len(possibleMoves))]
							rolloutMove(rolloutStatus, randomAction, player)
							countLivingPlayers++
						}
					}
					//All other players Died
					if countLivingPlayers == 0 {
						break
					}
				}
				possibleMoves := me.PossibleMoves(rolloutStatus.Cells, rolloutStatus.Turn, nil, false)
				if len(possibleMoves) == 0 {
					break
				}
				var randomAction Action
				//This should distribute the first Action taken equally
				if counter == 0 {
					randomAction = possibleMoves[performedRollouts%len(possibleMoves)]
					counter++
				} else {
					randomAction = possibleMoves[rand.Intn(len(possibleMoves))]
				}
				rolloutMove(rolloutStatus, randomAction, me)
				path = append(path, randomAction)
			}
			longestPaths, longest = checkPath(path, longestPaths, longest, performedRollouts)
		}
	}
	log.Println("could perfom", maxNumberofRollouts, "rollouts, which is the maximum possible")
	return longestPaths
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
func checkPath(path []Action, longestPaths [][]Action, longest int, allreadyPerformedRollouts int) ([][]Action, int) {
	//Now we chek if the last taken path was longer then every other path
	if float64(len(path)) >= float64(longest)*filterValue {
		//if longest is still bigger then the path found now we only append the path
		if longest >= len(path) {
			longestPaths = append(longestPaths, path)
			//If it is bigger by a lot we can forget every path we found until now
		} else if float64(len(path))*filterValue > float64(longest) {
			longestPaths = make([][]Action, 0, maxNumberofRollouts-allreadyPerformedRollouts)
			longestPaths = append(longestPaths, path)
			longest = len(path)
			//If none of the before is the case we have to filter all values that are in longest paths until now
		} else {
			longestPaths = filterPaths(longestPaths, len(path), filterValue, maxNumberofRollouts-allreadyPerformedRollouts+len(longestPaths))
			longestPaths = append(longestPaths, path)
			longest = len(path)
		}
	}
	return longestPaths, longest

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
