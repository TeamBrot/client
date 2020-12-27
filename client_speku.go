package main

import (
	"log"
	"math"
	"math/rand"
	"runtime"
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

//Checks if it is legal for a SimPlayer to use a Cell
func checkCell(cells [][]bool, direction Direction, y uint16, x uint16, fields uint16, visitedCells map[Coords]struct{}, miniMaxSwitch bool) bool {
	var isPossible bool
	if direction == Up {
		y -= fields
	} else if direction == Down {
		y += fields
	} else if direction == Left {
		x -= fields
	} else {
		x += fields
	}
	if x >= uint16(len(cells[0])) || y >= uint16(len(cells)) || x < 0 || y < 0 {
		return false
	}
	isPossible = !cells[y][x]
	if visitedCells != nil {
		coordsNow := Coords{y, x}
		_, fieldVisited := visitedCells[coordsNow]
		if miniMaxSwitch {
			return isPossible || fieldVisited
		}
		return isPossible && !fieldVisited

	}
	return isPossible
}

// Returns possible actions for a given situation for a SimPlayer
func possibleMoves(player *Player, cells [][]bool, turn uint16, visitedFields map[Coords]struct{}, miniMaxSwitch bool) []Action {
	changeNothing := true
	turnRight := true
	turnLeft := true
	slowDown := player.Speed != 1
	speedUp := player.Speed != 10
	direction := player.Direction
	y := player.Y
	x := player.X
	for i := uint16(1); i <= uint16(player.Speed); i++ {
		checkJump := turn%6 == 0 && i > 1 && i < uint16(player.Speed)
		checkJumpSlowDown := turn%6 == 0 && i > 1 && i < uint16(player.Speed)-1
		checkJumpSpeedUp := turn%6 == 0 && i > 1 && i <= uint16(player.Speed)

		turnLeft = turnLeft && (checkJump || checkCell(cells, (direction+1)%4, y, x, i, visitedFields, miniMaxSwitch))
		changeNothing = changeNothing && (checkJump || checkCell(cells, direction, y, x, i, visitedFields, miniMaxSwitch))
		turnRight = turnRight && (checkJump || checkCell(cells, (direction+3)%4, y, x, i, visitedFields, miniMaxSwitch))
		if i != uint16(player.Speed) {
			slowDown = slowDown && (checkJumpSlowDown || checkCell(cells, direction, y, x, i, visitedFields, miniMaxSwitch))
		}
		speedUp = speedUp && (checkJumpSpeedUp || checkCell(cells, direction, y, x, i, visitedFields, miniMaxSwitch))
	}
	speedUp = speedUp && checkCell(cells, direction, y, x, uint16(player.Speed+1), visitedFields, miniMaxSwitch)

	possibleMoves := make([]Action, 0, 5)

	if slowDown {
		possibleMoves = append(possibleMoves, SlowDown)
	}
	if changeNothing {
		possibleMoves = append(possibleMoves, ChangeNothing)
	}
	if speedUp {
		possibleMoves = append(possibleMoves, SpeedUp)
	}
	if turnLeft {
		possibleMoves = append(possibleMoves, TurnLeft)
	}
	if turnRight {
		possibleMoves = append(possibleMoves, TurnRight)
	}
	return possibleMoves
}

//Simulates a Action and for a simPlayer and a board. Raises the score of every visited cell at the board and adds the Coords to allVisitedCells and lastMoveVisitedCells
func simulateMove(board [][]uint16, parentPlayer *SimPlayer, action Action, turn uint16, simField [][]float64) (*SimPlayer, float64) {
	childPlayer := parentPlayer.copySimPlayer()
	player := childPlayer.player
	score := 0.0
	visitedCoords := player.processAction(action, turn)
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

func (player *Player) processAction(action Action, turn uint16) []*Coords {
	if action == SpeedUp {
		player.Speed++
	} else if action == SlowDown {
		player.Speed--
	} else if action == TurnLeft {
		switch player.Direction {
		case Left:
			player.Direction = Down
			break
		case Down:
			player.Direction = Right
			break
		case Right:
			player.Direction = Up
			break
		case Up:
			player.Direction = Left
			break
		}
	} else if action == TurnRight {
		switch player.Direction {
		case Left:
			player.Direction = Up
			break
		case Down:
			player.Direction = Left
			break
		case Right:
			player.Direction = Down
			break
		case Up:
			player.Direction = Right
			break
		}
	}
	visitedCoords := make([]*Coords, player.Speed+1)
	jump := turn%6 == 0
	for i := uint8(1); i <= player.Speed; i++ {
		if player.Direction == Up {
			player.Y--
		} else if player.Direction == Down {
			player.Y++
		} else if player.Direction == Right {
			player.X++
		} else if player.Direction == Left {
			player.X--
		}
		if !jump || i == 1 || i == player.Speed {
			visitedCoords[i] = &Coords{player.Y, player.X}
		}
	}
	return visitedCoords
}

//implements the doMove function for the rollout function
func rolloutMove(status *Status, action Action, player *Player) {
	visitedCoords := player.processAction(action, status.Turn)
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
			return longestPaths
		default:
			rolloutStatus := status.copyStatus()
			path := make([]Action, 0)
			for i := 0; i < limit; i++ {
				me := rolloutStatus.Players[status.You]
				countLivingPlayers := 0
				rolloutStatus.Turn++
				//Process one random move for every other player besides me
				for _, player := range rolloutStatus.Players {
					if player != me && player != nil {
						possibleMoves := possibleMoves(player, rolloutStatus.Cells, rolloutStatus.Turn, nil, false)
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
				possibleMoves := possibleMoves(me, rolloutStatus.Cells, rolloutStatus.Turn, nil, false)
				if len(possibleMoves) == 0 {
					break
				}
				var randomAction Action
				//This should distribute the first Action taken equally
				if i == 0 {
					randomAction = possibleMoves[j%len(possibleMoves)]
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
	log.Println("Could perfom ", maxNumberofRollouts, " Rollouts, which is the maximum possible")
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
		}
		allSimPlayer = append(allSimPlayer, player.toSimPlayer())
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
		results := make([]*Result, 0)
		for _, ch := range resultChannels {
			select {
			case <-stopSimulateGameChan:
				log.Println("Ended Simulate Game. Returning field")
				return allProbabilityTables
			case val := <-ch:
				if val != nil {
					results = append(results, val)
				} else {
					break
				}
			}
		}
		if len(results) == 0 {
			break
		}
		log.Println("Starting calculating field for turn: ", z)
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
				log.Println("Ended Simulation for player ", id)
				currentPlayers = nil
				children = nil
				resultChannel <- nil
				return
			default:
				possibleActions := possibleMoves(player.player, status.Cells, elapsedTurns+uint16(turn), player.AllVisitedCells, false)
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
	log.Println("Finished Calculation for player: ", id)
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
	visitedCoords := player.processAction(action, turn)
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
	var scoreNothing float64
	var scoreLeft float64
	var scoreRight float64
	var scoreSlow float64
	var scoreSpeed float64
	possibleNothing := false
	possibleLeft := false
	possibleRight := false
	possibleUp := false
	possibleDown := false
	//Computes if a action is possible based on the possibleActions Array
	for _, action := range possibleActions {
		switch action {
		case ChangeNothing:
			possibleNothing = true
		case TurnLeft:
			possibleLeft = true
		case TurnRight:
			possibleRight = true
		case SpeedUp:
			possibleUp = true
		case SlowDown:
			possibleDown = true
		}
	}
	turn++
	//computes the score for every path
	for _, path := range paths {
		score := 0.0
		newPlayer := player.copyPlayer()
		for i := 0; i < simDepth; i++ {
			if i != len(path) {
				score += evaluateAction(newPlayer, allFields[i], path[i], turn+uint16(i))
				score /= float64(simDepth)
			} else {
				break
			}
		}
		if len(path) == 0 {
			log.Println("All other players are going to die in the next turn")
			return possibleActions[0]
		}
		switch path[0] {
		case ChangeNothing:
			scoreNothing += score
		case TurnLeft:
			scoreLeft += score
		case TurnRight:
			scoreRight += score
		case SpeedUp:
			scoreSpeed += score
		case SlowDown:
			scoreSlow += score
		}
	}
	//computes how many times a Action was the first Action of path
	counterNothing := 1
	counterLeft := 1
	counterRight := 1
	counterUp := 1
	counterDown := 1
	for _, path := range paths {
		switch path[0] {
		case ChangeNothing:
			counterNothing++
		case TurnLeft:
			counterLeft++
		case TurnRight:
			counterRight++
		case SpeedUp:
			counterUp++
		case SlowDown:
			counterDown++
		}
	}
	//computes Value based on the score of a Action an
	valueNothing := (scoreNothing / float64(counterNothing)) + (1.0 - (float64(counterNothing) / float64(len(paths))))
	valueLeft := (scoreLeft / float64(counterLeft)) + (1.0 - (float64(counterLeft) / float64(len(paths))))
	valueRight := (scoreRight / float64(counterRight)) + (1.0 - (float64(counterRight) / float64(len(paths))))
	valueUp := (scoreSpeed / float64(counterUp)) + (1.0 - (float64(counterUp) / float64(len(paths))))
	valueDown := (scoreSlow / float64(counterDown)) + (1.0 - (float64(counterDown) / float64(len(paths))))
	log.Println("Calculated Scores")
	log.Println("Change Nothing: ", valueNothing)
	log.Println("Turn Left: ", valueLeft)
	log.Println("TurnRight: ", valueRight)
	log.Println("Speed Up: ", valueUp)
	log.Println("Slow Down: ", valueDown)
	if possibleNothing && (valueNothing < valueLeft || !possibleLeft) && (!possibleRight || valueNothing < valueRight) && (valueNothing < valueUp || !possibleUp) && (valueNothing < valueDown || !possibleDown) {
		return ChangeNothing
	} else if possibleLeft && (valueLeft < valueRight || !possibleRight) && (valueLeft < valueUp || !possibleUp) && (valueLeft < valueDown || !possibleDown) {
		return TurnLeft
	} else if possibleRight && (valueRight < valueUp || !possibleUp) && (valueRight < valueDown || !possibleDown) {
		return TurnRight
	} else if possibleUp && (valueUp < valueDown || !possibleDown) {
		return SpeedUp
	} else {
		return SlowDown
	}
}

//This Method is work in progress and does basically nothing
func analyzeBoard(status *Status) []*Player {
	var playerSimulation []*Player
	//var playerRollouts []Player
	numberOfCells := status.Width * status.Height

	var numberOfOccupiedCells float64
	for _, row := range status.Cells {
		for _, cellValue := range row {
			if cellValue {
				numberOfOccupiedCells++
			}
		}
	}
	boardCoverage := numberOfOccupiedCells / float64(numberOfCells)
	log.Println(boardCoverage*100, "% of the board are used")

	influenceWidhtOfPlayer := make(map[*Player]int, 0)

	for _, player := range status.Players {
		influenceWidth := (player.Speed - uint8(math.Round(9*boardCoverage)) + 4) * 8
		log.Println(influenceWidth)
		influenceWidhtOfPlayer[player] = int(influenceWidth)
	}
	me := status.Players[status.You]
	for _, player := range status.Players {
		if player == me {
			continue
		}
		//
		//if distanceToPlayer(me, player) < float64(influenceWidhtOfPlayer[player]+influenceWidhtOfPlayer[me]) {
		playerSimulation = append(playerSimulation, player)
		//}
	}
	playerSimulation = append(playerSimulation, me)
	return playerSimulation
}

func spekuTiming(calculationTime time.Duration, timingChannel chan<- time.Time) {
	time.Sleep(time.Duration(0.4 * float64(calculationTime.Nanoseconds())))
	timingChannel <- time.Now()
	time.Sleep(time.Duration(0.4 * float64(calculationTime.Nanoseconds())))
	close(timingChannel)
}

func (player *Player) toSimPlayer() *SimPlayer {
	var simPlayer SimPlayer
	simPlayer.player = player.copyPlayer()
	simPlayer.AllVisitedCells = make(map[Coords]struct{}, 0)
	simPlayer.LastMoveVisitedCells = make(map[Coords]struct{}, 0)
	simPlayer.Probability = 1.0
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
	possibleActions := possibleMoves(&player, status.Cells, status.Turn, nil, false)
	//handle trivial cases (zero or one possible Action)
	if len(possibleActions) == 1 {
		log.Println("Only possible Action: ", possibleActions[0])
		return possibleActions[0]
	} else if len(possibleActions) == 0 {
		log.Println("I'll die... choosing change nothing as last action")
		return ChangeNothing
	}

	otherPlayerID := findClosestPlayer(status)
	log.Println("using player", otherPlayerID, "at", status.Players[otherPlayerID].X, status.Players[otherPlayerID].Y, "as minimizer")
	miniMaxChannel := make(chan []Action, 1)
	stopMiniMaxChannel := make(chan time.Time)
	go func() {
		miniMaxActions := bestActionsMinimaxTimed(status.You, otherPlayerID, status, stopMiniMaxChannel)
		miniMaxChannel <- miniMaxActions
	}()
	stopRolloutChan := make(chan time.Time)
	rolloutChan := make(chan [][]Action, 1)
	go func() {
		rolloutPaths := simulateRollouts(status, 140, 0.7, stopRolloutChan)
		rolloutChan <- rolloutPaths
	}()

	//calculate which players are simulated TODO: Move this code to an external function and improve it
	activePlayersInRange := analyzeBoard(status)
	log.Println("Simulating ", len(activePlayersInRange), " Players")
	var maxSimDepth int
	//if Your Computer is really beefy it might be a good idea to set this higher (else it is not and your computer will crash!!)
	maxSimDepth = 12
	//If this channel is closed, it will try to end simulate game

	//This channel is used to recieve an array of all calculated Fields from simulate game
	var allProbabilityTables [][][]float64
	var probabilityTablesChan chan [][][]float64
	var stopSimulateGameChan chan time.Time
	if len(activePlayersInRange) > 1 {
		stopSimulateGameChan = make(chan time.Time)
		probabilityTablesChan = make(chan [][][]float64, 1)
		go func() {
			probabilityTables := simulateGame(status, stopSimulateGameChan, maxSimDepth, activePlayersInRange)
			probabilityTablesChan <- probabilityTables
		}()
	} else {
		allProbabilityTables = make([][][]float64, maxSimDepth)
		for z := range allProbabilityTables {
			allProbabilityTables[z] = makeProbabilityTable(status.Height, status.Width)
		}
	}

	_ = <-timingChannel
	if len(activePlayersInRange) > 1 {
		log.Println("Sending stop Signal to simulate Game...")
		close(stopSimulateGameChan)
	}
	_ = <-timingChannel
	log.Println("Sending stop signal to Simulate Rollouts and MiniMax...")
	close(stopRolloutChan)
	close(stopMiniMaxChannel)
	possibleActions = <-miniMaxChannel
	if len(activePlayersInRange) > 1 {
		allProbabilityTables = <-probabilityTablesChan
	}
	bestPaths := <-rolloutChan

	log.Println("Found ", len(bestPaths), " Paths that should be evaluated")
	log.Println("Could calculate ", len(allProbabilityTables), " Turns")
	//This is only for debugging purposes and combines the last field with the status
	log.Println(allProbabilityTables[len(allProbabilityTables)-1])
	//Log Timing
	log.Println("Time till Calculations are finished and Evaluation can start: ", time.Since(start))
	//Evaluate the paths with the given field and return the best Action based on this TODO: Needs improvement in case of naming
	bestAction = evaluatePaths(player, allProbabilityTables, bestPaths, status.Turn, len(allProbabilityTables)-1, possibleActions)
	//Log Timing
	log.Println("The Processing in total took: ", time.Since(start))
	log.Println("Choose as best action: ", bestAction)
	return bestAction
}
