package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

type Result struct {
	Cells  [][]int
	Player []*SimPlayer
	turn   int
	ID     int
}
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
	probability          float64
	allVisitedCells      map[Coords]struct{}
	lastMoveVisitedCells map[Coords]struct{}
	parent               *SimPlayer
}

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
	_, fieldVisited := player.allVisitedCells[coordsNow]
	return cells[y][x] == 0 && !fieldVisited
}

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

	possibleMoves := make([]Action, 0)

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

//Simulates a move. Doesnt mark a field as occupied. Only writes the probability, that it could be occupied. (Needs improvment at the undo part)
func simulateMove(fieldPointer *[][]int, parentPlayer *SimPlayer, action Action, turn int) *SimPlayer {
	field := *fieldPointer
	player := copyPlayer(parentPlayer)
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
			field[player.Y][player.X]++
			coordsNow := Coords{player.Y, player.X}
			player.allVisitedCells[coordsNow] = struct{}{}
			player.lastMoveVisitedCells[coordsNow] = struct{}{}
		}
	}
	return player
}

//implements the doMove function for the rollout function (it is possible to take illegal moves -> the player dies)
func rolloutMove(status *Status, action Action) {
	player := status.Players[status.You]
	turn := status.Turn
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

			status.Cells[player.Y][player.X] = status.You

		}
	}

}

//searche for the longest paths a player could reach. Only simulates moves for one player!
func simulateRollouts(status *Status, limit int, ch chan [][]Action) [][]Action {
	longest := 0
	filterValue := 0.8
	longestPaths := make([][]Action, 0)
	for j := 0; j < 20000; j++ {
		rolloutStatus := copyStatus(status)
		path := make([]Action, 0)
		for i := 0; i < limit; i++ {
			possibleMoves := Moves(rolloutStatus, rolloutStatus.Players[status.You], nil)
			if len(possibleMoves) == 0 {
				break
			}

			rolloutStatus.Turn = rolloutStatus.Turn + 1
			randomAction := possibleMoves[rand.Intn(len(possibleMoves))]
			rolloutMove(rolloutStatus, randomAction)
			path = append(path, randomAction)
		}
		if float64(len(path)) >= float64(longest)*filterValue {

			if longest >= len(path) {
				longestPaths = append(longestPaths, path)
			} else {
				longestPaths = filterPaths(longestPaths, len(path), filterValue)
				longestPaths = append(longestPaths, path)
				longest = len(path)
			}
		}
	}
	if ch != nil {
		ch <- longestPaths
	}
	return longestPaths

}

func filterPaths(paths [][]Action, longest int, percent float64) [][]Action {
	filteredPaths := make([][]Action, 0)
	for _, path := range paths {
		if float64(len(path)) >= float64(longest)*percent {
			filteredPaths = append(filteredPaths, path)
		}
	}
	return filteredPaths
}

//Copys a SimPlayer Object (Might be transfered to a util.go)
func copyPlayer(player *SimPlayer) *SimPlayer {
	var p SimPlayer
	p.Direction = player.Direction
	p.Speed = player.Speed
	p.X = player.X
	p.Y = player.Y
	p.parent = player
	p.lastMoveVisitedCells = make(map[Coords]struct{})
	p.allVisitedCells = make(map[Coords]struct{})
	for k := range player.allVisitedCells {
		p.allVisitedCells[k] = struct{}{}
	}
	return &p
}

//TODO: simulate game
func simulateGame(status *Status, chField chan [][]float64, numberOfTurns int) {
	allSimPlayer := make([]*SimPlayer, 0)
	for _, player := range status.Players {
		if player.Active && status.Players[status.You] != player {
			var p SimPlayer
			p.Direction = player.Direction
			p.Speed = player.Speed
			p.X = player.X
			p.Y = player.Y
			p.parent = nil
			p.lastMoveVisitedCells = make(map[Coords]struct{})
			p.allVisitedCells = make(map[Coords]struct{})
			allSimPlayer = append(allSimPlayer, &p)
		}
	}
	resultChannels := make(map[int]chan Result, 0)
	for j, simPlayer := range allSimPlayer {
		playerTree := make([][]*SimPlayer, 1)
		playerTree[0] = make([]*SimPlayer, 1)
		playerTree[0][0] = simPlayer
		resultChannels[j] = make(chan Result, numberOfTurns)
		go simulatePlayer(playerTree, j, status, numberOfTurns, status.Turn, resultChannels[j])
	}
	// combine fields
	for z := 0; z < numberOfTurns; z++ {
		results := make([]*Result, 0)
		for _, ch := range resultChannels {
			val := <-ch
			results = append(results, &val)
		}
		go resultsToField(results, status.Width, status.Height, chField)
	}
	// calculate Malus with go routines

	// calculate probabilities and return field
	return
}

func resultsToField(results []*Result, width int, height int, ch chan [][]float64) {
	intermediateFields := make([][][]int, len(results))
	for u := 0; u < len(results); u++ {
		intermediateFields[u] = makeEmptyCells(height, width)
		//fmt.Println("the intermediate field")
		//fmt.Println(intermediateFields[u])
	}
	for m, cells := range intermediateFields {
		for n, result := range results {
			if n != m {
				cells = addCells(cells, result.Cells)
			}
		}
	}

	for g, cells := range intermediateFields {
		result := results[g]
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if cells[y][x] != 0 {
					coordsNow := Coords{y, x}
					for _, player := range result.Player {
						_, isIn := player.lastMoveVisitedCells[coordsNow]
						if isIn {
							player.probability = player.probability - 1.0/float64(cells[y][x])
						}
					}
				}
			}
		}
	}
	resultField := make([][]float64, height)
	for g := range resultField {
		resultField[g] = make([]float64, width)
	}
	for _, result := range results {
		for _, player := range result.Player {
			player.probability = player.probability * player.parent.probability
			for coords := range player.lastMoveVisitedCells {
				resultField[coords.Y][coords.X] = resultField[coords.Y][coords.X] + player.probability
			}
		}
	}

	ch <- resultField

	return

}

// simulates all possible moves for a given field.
func simulatePlayer(playerTree [][]*SimPlayer, id int, status *Status, numberOfTurns int, elapsedTurns int, ch chan Result) {
	for i := 0; i < numberOfTurns; i++ {
		playerTree = append(playerTree, make([]*SimPlayer, 0))
		turn := i + 1
		writeField := make([][]int, status.Height)
		for r := range writeField {
			writeField[r] = make([]int, status.Width)
		}
		for _, player := range playerTree[i] {
			children := make([]*SimPlayer, 0)
			for _, action := range possibleMoves(player, status.Cells, elapsedTurns+turn) {
				newPlayer := simulateMove(&writeField, player, action, elapsedTurns+turn)
				children = append(children, newPlayer)
			}
			for _, child := range children {
				child.probability = 1.0 / float64(len(children))
				playerTree[turn] = append(playerTree[turn], child)
			}
		}
		if ch != nil {
			ch <- Result{Cells: writeField, ID: id, Player: playerTree[turn], turn: turn}
		}
	}
	close(ch)
	return
}

func addFields(field1 [][]float64, field2 [][]float64) [][]float64 {
	for i := 0; i < len(field1); i++ {
		for j := 0; j < len(field1[i]); j++ {
			field2[i][j] = field2[i][j] + field1[i][j]
		}
	}
	return field2
}

func addCells(field1 [][]int, field2 [][]int) [][]int {
	for i := 0; i < len(field1); i++ {
		for j := 0; j < len(field1[i]); j++ {
			field2[i][j] = field2[i][j] + field1[i][j]
		}
	}
	return field2
}

// returns an empty field
func makeEmptyCells(height int, width int) [][]int {
	field := make([][]int, height)
	for r := range field {
		field[r] = make([]int, width)
	}
	return field
}

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
func copyNormalPlayer(oldPlayer *Player) *Player {
	var newPlayer Player
	newPlayer.Direction = oldPlayer.Direction
	newPlayer.Speed = oldPlayer.Speed
	newPlayer.X = oldPlayer.X
	newPlayer.Y = oldPlayer.Y
	return &newPlayer
}
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
	for _, path := range paths {
		score := 0.0
		newPlayer := copyNormalPlayer(&player)
		for i := 0; i < simDepth; i++ {
			if i != len(path) {
				score += evaluateScore(newPlayer, allFields[i], path[i], turn+i)
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
	valueNothing := scoreNothing + (float64(len(paths)) / float64(counterNothing))
	valueLeft := scoreLeft + (float64(len(paths)) / float64(counterLeft))
	valueRight := scoreRight + (float64(len(paths)) / float64(counterRight))
	valueUp := scoreSpeed + (float64(len(paths)) / float64(counterUp))
	valueDown := scoreSlow + (float64(len(paths)) / float64(counterDown))

	fmt.Println("Change Nothing: ", valueNothing)
	fmt.Println("Turn Left: ", valueLeft)
	fmt.Println("TurnRight: ", valueRight)
	fmt.Println("Speed Up: ", valueUp)
	fmt.Println("Slow Down: ", valueDown)
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

// SpekuClient is a client implementation that uses speculation to decide what to do next
type SpekuClient struct{}

// GetAction implements the Client interface
//TODO: use player information
func (c SpekuClient) GetAction(player Player, status *Status) Action {
	start := time.Now()
	var bestAction Action
	possibleActions := Moves(status, &player, nil)
	if len(possibleActions) == 1 {
		fmt.Println("Only possible Action: ", possibleActions[0])
		return possibleActions[0]
	} else if len(possibleActions) == 0 {
		fmt.Println("I'll die...")
		return ChangeNothing
	}
	simChan := make(chan [][]Action, 1)
	go simulateRollouts(status, 10000, simChan)
	simDepth := 9
	fieldChan := make(chan [][]float64, simDepth)
	go simulateGame(status, fieldChan, simDepth)
	otherPlayerID := findClosestPlayer(status)
	log.Println("using player", otherPlayerID, "at", status.Players[otherPlayerID].X, status.Players[otherPlayerID].Y, "as minimizer")
	possibleActions = bestActionsMinimax(status.You, otherPlayerID, status, 6)
	allFields := make([][][]float64, simDepth+1)
	for i := 0; i < simDepth; i++ {
		allFields[i] = <-fieldChan
	}
	bestPaths := <-simChan
	bestAction = evaluatePaths(player, allFields, bestPaths, status.Turn, simDepth, possibleActions)
	fmt.Println(len(bestPaths))

	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println(elapsed)
	fmt.Println(bestAction)
	return bestAction
}
