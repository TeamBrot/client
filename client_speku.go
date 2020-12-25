package main

import (
	"log"
	"math"
	"math/rand"
	"time"
)

// Result :
type Result struct {
	Cells  [][]int
	Player []*SimPlayer
	turn   int
	ID     int
}

// EvaluateStatus :
type EvaluateStatus struct {
	Width    int           `json:"width"`
	Height   int           `json:"height"`
	AllCells [][][]float64 `json:"cells"`
	Player   Player
	Turn     int
}

//Coords store the coordinates of a player
type Coords struct {
	Y, X int
}

//SimPlayer to add a new array of visited cells
type SimPlayer struct {
	X                    int `json:"x"`
	Y                    int `json:"y"`
	Direction            Direction
	Speed                int `json:"speed"`
	Probability          float64
	AllVisitedCells      map[Coords]struct{}
	LastMoveVisitedCells map[Coords]struct{}
	Parent               *SimPlayer
}

//Checks if it is legal for a SimPlayer to use a Cell
func checkCell2(cells [][]int, direction Direction, y int, x int, fields int, player *SimPlayer) bool {
	if direction == Up {
		y -= fields
	} else if direction == Down {
		y += fields
	} else if direction == Left {
		x -= fields
	} else {
		x += fields
	}
	if x >= len(cells[0]) || y >= len(cells) || x < 0 || y < 0 {
		return false
	}
	coordsNow := Coords{y, x}
	_, fieldVisited := player.AllVisitedCells[coordsNow]
	return cells[y][x] == 0 && !fieldVisited
}

// Returns possible actions for a given situation for a SimPlayer
func possibleMoves(player *SimPlayer, cells [][]int, turn int) []Action {
	changeNothing := true
	turnRight := true
	turnLeft := true
	slowDown := player.Speed != 1
	speedUp := player.Speed != 10
	direction := player.Direction
	y := player.Y
	x := player.X
	for i := 1; i <= player.Speed; i++ {
		checkJump := turn%6 == 0 && i > 1 && i < player.Speed
		checkJumpSlowDown := turn%6 == 0 && i > 1 && i < player.Speed-1
		checkJumpSpeedUp := turn%6 == 0 && i > 1 && i <= player.Speed

		turnLeft = turnLeft && (checkJump || checkCell2(cells, (direction+1)%4, y, x, i, player))
		changeNothing = changeNothing && (checkJump || checkCell2(cells, direction, y, x, i, player))
		turnRight = turnRight && (checkJump || checkCell2(cells, (direction+3)%4, y, x, i, player))
		if i != player.Speed {
			slowDown = slowDown && (checkJumpSlowDown || checkCell2(cells, direction, y, x, i, player))
		}
		speedUp = speedUp && (checkJumpSpeedUp || checkCell2(cells, direction, y, x, i, player))
	}
	speedUp = speedUp && checkCell2(cells, direction, y, x, player.Speed+1, player)

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
func simulateMove(boardPointer *[][]int, parentPlayer *SimPlayer, action Action, turn int, simField [][]float64) (*SimPlayer, float64) {
	board := *boardPointer
	player := copySimPlayer(parentPlayer)
	score := 0.0
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

	jump := turn%6 == 0
	for i := 1; i <= player.Speed; i++ {
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
			board[player.Y][player.X]++
			coordsNow := Coords{player.Y, player.X}
			player.AllVisitedCells[coordsNow] = struct{}{}
			player.LastMoveVisitedCells[coordsNow] = struct{}{}
			score += simField[player.Y][player.X]
		}
	}
	return player, score
}

//implements the doMove function for the rollout function
func rolloutMove(status *Status, action Action, player *Player) {
	turn := status.Turn
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

	jump := turn%6 == 0
	for i := 1; i <= player.Speed; i++ {
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
			status.Cells[player.Y][player.X] = status.You
		}
	}

}

//search for the longest paths a player could reach. Simulates random move for all Players and allways processes as last player
func simulateRollouts(status *Status, limit int, filterValue float64, ch chan [][]Action, stopSimulateRollouts <-chan time.Time) {
	longest := 0
	maxNumberofRollouts := 7000000
	longestPaths := make([][]Action, 0, maxNumberofRollouts)
	for j := 0; j < maxNumberofRollouts; j++ {
		select {
		case <-stopSimulateRollouts:
			ch <- longestPaths
			log.Println("Could perfomr ", j, " Rollouts")
			close(ch)
			return
		default:
			rolloutStatus := copyStatus(status)
			path := make([]Action, 0)
			for i := 0; i < limit; i++ {
				me := rolloutStatus.Players[status.You]
				counter := 0
				for _, player := range rolloutStatus.Players {
					if player != me && player.Active {
						possibleMoves := Moves(rolloutStatus, player, nil)
						if len(possibleMoves) == 0 {
							player.Active = false
							continue
						}
						randomAction := possibleMoves[rand.Intn(len(possibleMoves))]
						rolloutMove(rolloutStatus, randomAction, player)
						counter++
					}
				}
				if counter == 0 {
					break
				}
				possibleMoves := Moves(rolloutStatus, me, nil)
				if len(possibleMoves) == 0 {
					break
				}
				rolloutStatus.Turn = rolloutStatus.Turn + 1
				var randomAction Action
				if i == 0 {
					randomAction = possibleMoves[j%len(possibleMoves)]
				} else {
					randomAction = possibleMoves[rand.Intn(len(possibleMoves))]
				}
				rolloutMove(rolloutStatus, randomAction, me)
				path = append(path, randomAction)
			}
			if float64(len(path)) >= float64(longest)*filterValue {

				if longest >= len(path) {
					longestPaths = append(longestPaths, path)
				} else if float64(len(path))*filterValue > float64(longest) {
					longestPaths = make([][]Action, 0, maxNumberofRollouts-j)
					longestPaths = append(longestPaths, path)
					longest = len(path)
				} else {
					longestPaths = filterPaths(longestPaths, len(path), filterValue, maxNumberofRollouts-j+len(longestPaths))
					longestPaths = append(longestPaths, path)
					longest = len(path)
				}
			}
		}
	}
	ch <- longestPaths
	log.Println("Could perfomr ", maxNumberofRollouts, " Rollouts")
	return
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
func copySimPlayer(player *SimPlayer) *SimPlayer {
	var p SimPlayer
	p.Direction = player.Direction
	p.Speed = player.Speed
	p.X = player.X
	p.Y = player.Y
	p.Parent = player
	p.LastMoveVisitedCells = make(map[Coords]struct{})
	p.AllVisitedCells = make(map[Coords]struct{})
	for k := range player.AllVisitedCells {
		p.AllVisitedCells[k] = struct{}{}
	}
	return &p
}

// Simulate games for all given Players and the given SimDepth. Uses simulatePlayer and resultsToField to achieve this
func simulateGame(status *Status, chField chan<- [][][]float64, stopSimulateGameChan <-chan time.Time, simDepth int, activePlayersInRange []*Player) {
	var me int
	allSimPlayer := make([]*SimPlayer, 0)
	//Convert all Players that should be simulated to simPlayers
	for d, player := range activePlayersInRange {
		if player == status.Players[status.You] {
			me = d
		}
		var p SimPlayer
		p.Direction = player.Direction
		p.Speed = player.Speed
		p.Probability = 1
		p.X = player.X
		p.Y = player.Y
		p.Parent = nil
		p.LastMoveVisitedCells = make(map[Coords]struct{})
		p.AllVisitedCells = make(map[Coords]struct{})
		allSimPlayer = append(allSimPlayer, &p)
	}
	//with those Channels resultsToField can give fields to simulatePlayer to make the calculations more acurate
	boardChannels := make(map[int]chan [][]float64, 0)
	//with those Channels simulatePlayer gives a result back to simulateGame to process the result
	resultChannels := make(map[int]chan *Result, 0)
	for j, simPlayer := range allSimPlayer {

		boardChannels[j] = make(chan [][]float64, 1)
		resultChannels[j] = make(chan *Result, simDepth)
		go simulatePlayer(simPlayer, j, status, simDepth, status.Turn, resultChannels[j], boardChannels[j], stopSimulateGameChan)
	}
	allFields := make([][][]float64, 0)
	for z := 0; z < simDepth; z++ {
		results := make([]*Result, 0)
		for _, ch := range resultChannels {
			select {
			case <-stopSimulateGameChan:
				log.Println("Ended Simulate Game. Returning field")
				chField <- allFields
				return
			case val := <-ch:
				results = append(results, val)
			}

		}
		log.Println("Starting calculating field for turn: ", z)
		field := resultsToField(me, results, status.Width, status.Height, boardChannels)
		allFields = append(allFields, field)
		if z != 0 {
			addFields(&allFields[z], allFields[z-1])
		}
	}
	chField <- allFields
	return
}

//After every turn the given results are evaluated and fields are computed on basis of them
func resultsToField(me int, results []*Result, width int, height int, fieldChannels map[int]chan [][]float64) [][]float64 {
	storeBoards := make([][][]int, len(results))
	for u := 0; u < len(results); u++ {
		storeBoards[u] = makeEmptyBoard(height, width)
	}

	for m, cells := range storeBoards {
		for n, result := range results {
			if n != m {
				addBoards(&cells, result.Cells)
			}
		}
	}

	for g, cells := range storeBoards {
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

	playerFields := make([][][]float64, len(results))
	for k := range playerFields {
		newField := make([][]float64, height)
		for g := range newField {
			newField[g] = make([]float64, width)
		}
		playerFields[k] = newField
	}

	for l, result := range results {
		for _, player := range result.Player {
			if player == nil {
				break
			}
			player.Probability = player.Probability * player.Parent.Probability
			for coords := range player.LastMoveVisitedCells {
				for z, field := range playerFields {
					if z != l {
						field[coords.Y][coords.X] += player.Probability
					}
				}
			}
			player.Parent = nil
		}
	}
	for o, ch := range fieldChannels {
		ch <- playerFields[o]
	}
	return playerFields[me]

}

// simulate all possible moves for a given simPlayer and a given status until numberOfTurns is reached
func simulatePlayer(simPlayer *SimPlayer, id int, status *Status, numberOfTurns int, elapsedTurns int, resultChannel chan<- *Result, boardChannel <-chan [][]float64, stopChannel <-chan time.Time) {
	var fieldAfterTurn [][]float64
	playerTree := make([][]*SimPlayer, numberOfTurns+1)
	playerTree[0] = make([]*SimPlayer, 1)
	playerTree[0][0] = simPlayer
	for i := 0; i < numberOfTurns; i++ {
		playerTree[i+1] = make([]*SimPlayer, len(playerTree[i])*5)
		turn := i + 1
		writeField := make([][]int, status.Height)
		for r := range writeField {
			writeField[r] = make([]int, status.Width)
		}
		if i != 0 {
			select {
			case newField := <-boardChannel:
				addFields(&fieldAfterTurn, newField)
				playerTree[i-1] = make([]*SimPlayer, 0)
			case <-stopChannel:
				playerTree = make([][]*SimPlayer, 0)
				resultChannel <- nil
				return
			}

		} else {
			fieldAfterTurn = make([][]float64, status.Height)
			for z := range fieldAfterTurn {
				fieldAfterTurn[z] = make([]float64, status.Width)
			}
		}
		simPlayerCounter := 0
		for _, playerTreeTurn := range playerTree {
			simPlayerCounter += len(playerTreeTurn)
		}
		log.Println("Have to remember ", simPlayerCounter, " SimPlayers")
		counter := 0
		for _, player := range playerTree[i] {
			select {
			case <-stopChannel:
				log.Println("Ended Simulation for player ", id)
				playerTree = make([][]*SimPlayer, 0)
				resultChannel <- nil
				return
			default:
				possibleActions := possibleMoves(player, status.Cells, elapsedTurns+turn)
				for _, action := range possibleActions {
					child, score := simulateMove(&writeField, player, action, elapsedTurns+turn, fieldAfterTurn)
					child.Probability = 1.0/float64(len(possibleActions)) - (score / float64(len(child.LastMoveVisitedCells)))
					if child.Probability < 0 {
						continue
					}
					playerTree[turn][counter] = child
					counter++
				}

			}

		}
		playerTree[turn] = playerTree[turn][0:counter]
		if resultChannel != nil {
			resultChannel <- &Result{Cells: writeField, ID: id, Player: playerTree[turn], turn: turn}
		}
	}
	log.Println("Finished Calculation for player: ", id)
	playerTree = make([][]*SimPlayer, 0)
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

//Adds second board to first board. First board has to be a pointer an is going to be changed!
func addBoards(field1 *[][]int, field2 [][]int) {
	field := *field1
	for i := 0; i < len(field); i++ {
		for j := 0; j < len(field[i]); j++ {
			field[i][j] += field2[i][j]
		}
	}

}

// returns an empty field
func makeEmptyBoard(height int, width int) [][]int {
	field := make([][]int, height)
	for r := range field {
		field[r] = make([]int, width)
	}
	return field
}

//This functions executes a action and returns the average score of every visited Cell
func evaluateScore(player *Player, field [][]float64, action Action, turn int) float64 {
	score := 0.0
	if action == SpeedUp {
		if player.Speed != 10 {
			player.Speed++
		}
	} else if action == SlowDown {
		if player.Speed != 1 {
			player.Speed--
		}
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

	jump := turn%6 == 0
	for i := 1; i <= player.Speed; i++ {
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
			score += field[player.Y][player.X]
		}
	}
	return score / float64(player.Speed)
}

//This function copies a struct of type Player
func copyPlayer(oldPlayer *Player) *Player {
	var newPlayer Player
	newPlayer.Direction = oldPlayer.Direction
	newPlayer.Speed = oldPlayer.Speed
	newPlayer.X = oldPlayer.X
	newPlayer.Y = oldPlayer.Y
	return &newPlayer
}

//Simulates the moves of all Longest Paths until simDepth is reached. Computes a score for every possible Action and returns the Action with the lowes score
func evaluatePaths(player Player, allFields [][][]float64, paths [][]Action, turn int, simDepth int, possibleActions []Action) Action {
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
		newPlayer := copyPlayer(&player)
		for i := 0; i < simDepth; i++ {
			if i != len(path) {
				score += evaluateScore(newPlayer, allFields[i], path[i], turn+i)
				score /= float64(simDepth)
			} else {
				break
			}
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

//Function for debug purposes that combines a Field and a board to Visualize the output
func addFieldAndOriginalCells(field [][]float64, cells [][]int) [][]float64 {
	for i := 0; i < len(field); i++ {
		for j := 0; j < len(field[0]); j++ {
			field[i][j] += float64(cells[i][j])
		}
	}
	return field
}

//This Method is work in progress and does basically nothing
func analyzeBoard(status *Status) []*Player {
	//var playerSimulation []Player
	//var playerRollouts []Player
	numberOfCells := status.Width * status.Height

	var numberOfOccupiedCells float64
	for _, row := range status.Cells {
		for _, cellValue := range row {
			if cellValue != 0 {
				numberOfOccupiedCells++
			}
		}
	}
	boardCoverage := numberOfOccupiedCells / float64(numberOfCells)
	log.Println(boardCoverage, "% of the board are used")

	allActivePlayers := make([]*Player, 0)
	for _, player := range status.Players {
		if player.Active {
			allActivePlayers = append(allActivePlayers, player)
		}
	}

	influenceWidhtOfPlayer := make(map[*Player]int, 0)

	for _, player := range allActivePlayers {

		influenceWidth := player.Speed - int(math.Round(math.Pow(9, boardCoverage))) + 5
		log.Println(influenceWidth)
		influenceWidhtOfPlayer[player] = influenceWidth
	}
	return allActivePlayers
}

// SpekuClient is a client implementation that uses speculation to decide what to do next
type SpekuClient struct{}

// GetAction implements the Client interface
func (c SpekuClient) GetAction(player Player, status *Status, timingChannel <-chan time.Time) Action {
	start := time.Now()
	var bestAction Action
	possibleActions := Moves(status, &player, nil)
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
	possibleActions, _ = bestActionsMinimax(status.You, otherPlayerID, status, 3, nil)
	stopRolloutChan := make(chan time.Time)
	rolloutChan := make(chan [][]Action, 1)
	go simulateRollouts(status, 75, 0.7, rolloutChan, stopRolloutChan)

	//calculate which players are simulated TODO: Move this code to an external function and improve it
	activePlayersInRange := analyzeBoard(status)
	var maxSimDepth int
	//if Your Computer is really beefy it might be a good idea to set this higher (else it is not and your computer will crash!!)
	maxSimDepth = 9
	//If this channel is closed, it will try to end simulate game
	stopSimulateGameChan := make(chan time.Time)
	//This channel is used to recieve an array of all calculated Fields from simulate game
	fieldChan := make(chan [][][]float64, 1)
	go simulateGame(status, fieldChan, stopSimulateGameChan, maxSimDepth, activePlayersInRange)
	_ = <-timingChannel
	log.Println("Sending stop Signal to simulate Game...")
	close(stopSimulateGameChan)
	_ = <-timingChannel
	log.Println("Sending stop signal to Simulate Rollouts...")
	close(stopRolloutChan)
	allFields := <-fieldChan
	bestPaths := <-rolloutChan

	log.Println("Found ", len(bestPaths), " Paths that should be evaluated")
	log.Println("Could calculate ", len(allFields), " Turns")
	//This is only for debugging purposes and combines the last field with the status
	//log.Println(addFieldAndOriginalCells(allFields[len(allFields)-1], status.Cells))
	//Log Timing
	t2 := time.Now()
	elapsed2 := t2.Sub(start)
	log.Println("Time till Calculations are finished and Evaluation can start: ", elapsed2)

	//Evaluate the paths with the given field and return the best Action based on this TODO: Needs improvement in case of naming
	bestAction = evaluatePaths(player, allFields, bestPaths, status.Turn, len(allFields)-1, possibleActions)

	//Log Timing
	t3 := time.Now()
	elapsed3 := t3.Sub(start)
	log.Println("The Processing in total took: ", elapsed3)
	log.Println("Choose as best action: ", bestAction)
	return bestAction
}
