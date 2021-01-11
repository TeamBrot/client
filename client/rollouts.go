package main

import (
	"log"
	"math/rand"
	"time"
)

var validPathsToCache [][]Action

//If this value is set to true we process in every rollout before we choose our own action a action for every other living player
var simulateOtherPlayers = false

//This const defines the max number of Rollouts simulateRollouts will perform. Normally there is no good reason to change this value
const maxNumberofRollouts = 10000000

//search for the longest paths a player could reach. Simulates random action for all Players and allways processes as last player
func simulateRollouts(status *Status, stopSimulateRollouts <-chan time.Time, cachedPaths [][]Action, filterValue float64) [][]Action {
	var longestPaths [][]Action
	var longest int
	if cachedPaths != nil && !simulateOtherPlayers {
		longestPaths, longest = validateCachedPaths(cachedPaths, status)
		longestPaths = filterPaths(longestPaths, longest, filterValue)
		log.Println("Paths that are valid after filtering", len(longestPaths))
	} else {
		longest = 0
		longestPaths = make([][]Action, 0)
	}
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
				if simulateOtherPlayers {
					//Process one random action for every other player besides me
					for _, player := range rolloutStatus.Players {
						if player != me && player != nil {
							possibleActions := player.PossibleActions(rolloutStatus.Cells, rolloutStatus.Turn, nil, false)
							if len(possibleActions) == 0 {
								player = nil
								continue
							}
							randomAction := possibleActions[rand.Intn(len(possibleActions))]
							rolloutMove(rolloutStatus, randomAction, player)
						}
					}
				}
				possibleActions := me.PossibleActions(rolloutStatus.Cells, rolloutStatus.Turn, nil, false)
				if len(possibleActions) == 0 {
					break
				}
				var randomAction Action
				//This should distribute the first Action taken equally
				if counter == 0 {
					randomAction = possibleActions[performedRollouts%len(possibleActions)]
					counter++
				} else {
					randomAction = possibleActions[rand.Intn(len(possibleActions))]
				}
				rolloutMove(rolloutStatus, randomAction, me)
				rolloutStatus.Turn++
				path = append(path, randomAction)
			}
			longestPaths, longest = checkPath(path, longestPaths, longest, performedRollouts, filterValue)
			if len(longestPaths) > 5000 {
				longestPaths = filterPaths(longestPaths, longest, 0.9)
				log.Println("filter the longest paths cause there are too many of them, after filtering", len(longestPaths), "remaining")
				if len(longestPaths) > 100 {
					longestPaths = longestPaths[0:int(float64(len(longestPaths))/5)]
				}
				log.Println("keeping", len(longestPaths), "paths after all")
			}
		}
	}
	log.Println("could perfom", maxNumberofRollouts, "rollouts, which is the maximum possible")
	return longestPaths
}

//Checks if a Action, that was taken in the last turn is still valid
func checkIfActionIsPossible(checkedAction Action, status *Status, player *Player) bool {
	possibleActions := player.PossibleActions(status.Cells, status.Turn, nil, false)
	for _, action := range possibleActions {
		if action == checkedAction {
			return true
		}
	}
	return false
}

//Takes an Array of paths and checks if they are still valid and returns the array and the number of the longest path
func validateCachedPaths(oldPaths [][]Action, status *Status) ([][]Action, int) {
	newPaths := make([][]Action, 0)
	log.Println("Paths of the last round that are checked", len(oldPaths))
	longest := 0
	for _, path := range oldPaths {
		rolloutstatus := status.Copy()
		me := rolloutstatus.Players[rolloutstatus.You]
		lengthCounter := 0
		for _, action := range path {
			if checkIfActionIsPossible(action, rolloutstatus, me) {
				rolloutMove(rolloutstatus, action, me)
				lengthCounter++
				rolloutstatus.Turn++
			} else {
				break
			}
		}
		if lengthCounter > longest {
			longest = lengthCounter
		}
		newPaths = append(newPaths, path[0:lengthCounter])
	}
	log.Println("Length of the longest valid path of the last round", longest)
	return newPaths, longest
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

func checkPath(path []Action, longestPaths [][]Action, longest int, allreadyPerformedRollouts int, filterValue float64) ([][]Action, int) {
	//Now we chek if the last taken path was longer then every other path
	if float64(len(path)) >= float64(longest)*filterValue {
		//if longest is still bigger then the path found now we only append the path
		if longest >= len(path) {
			longestPaths = append(longestPaths, path)
		//If it is bigger by a lot we can forget every path we found until now
		} else if float64(len(path))*filterValue > float64(longest) {
			longestPaths = make([][]Action, 0)
			longestPaths = append(longestPaths, path)
			longest = len(path)
		//If none of the before is the case we have to filter all values that are in longest paths until now
		} else {
			longestPaths = filterPaths(longestPaths, len(path), filterValue)
			longestPaths = append(longestPaths, path)
			longest = len(path)
		}
	}
	return longestPaths, longest

}

//Filters an given array of paths and returns an array of paths that match the criteria
func filterPaths(paths [][]Action, longest int, percent float64) [][]Action {
	filteredPaths := make([][]Action, 0)
	for _, path := range paths {
		if float64(len(path)) >= float64(longest)*percent {
			filteredPaths = append(filteredPaths, path)
		}
	}
	return filteredPaths
}

//RolloutClient is a client implementation that uses only rollouts to decide what to do next
type RolloutClient struct {
	filterValue float64
}

// GetAction implements the Client interface
func (c RolloutClient) GetAction(status *Status, calculationTime time.Duration) Action {
	stopChannel := time.After((calculationTime / 10) * 9)
	stillValidPaths := validPathsToCache
	bestPaths := simulateRollouts(status, stopChannel, stillValidPaths, c.filterValue)
	possibleActions := status.Players[status.You].PossibleActions(status.Cells, status.Turn, nil, false)
	if len(possibleActions) == 0 {
		log.Println("I'll die")
		return ChangeNothing
	}
	var possible [5]bool
	//Computes if a action is possible based on the possibleActions Array
	for _, action := range possibleActions {
		possible[action] = true
	}

	counter := [5]int{1, 1, 1, 1, 1}
	for _, path := range bestPaths {
		counter[path[0]]++
	}
	var values [5]float64
	for i := 0; i < 5; i++ {
		values[i] = float64(counter[i]) / float64(len(bestPaths))
	}
	log.Println("calculated values", values)

	maximum := 0.0
	action := ChangeNothing
	for i, v := range values {
		if possible[i] && v > maximum {
			maximum = v
			action = Action(i)
		}
	}
	validPathsToCache = make([][]Action, 0)
	for _, path := range bestPaths {
		if len(path) > 1 {
			if path[0] == action {
				validPathsToCache = append(validPathsToCache, path[1:len(path)-1])
			}
		}
	}
	return action
}
