package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Result struct {
	Cells  [][]int
	Player []*SimPlayer
	turn   int
	ID     int
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

//Converts the Cells array of a status to a field
func convertCellsToField(status *Status) {
	cells := status.Cells
	height := status.Height
	width := status.Width
	for i, player := range status.Players {

		if i != status.You {
			fieldCells := make([][]float64, height)
			for i := range fieldCells {
				fieldCells[i] = make([]float64, width)
			}
			for i := range cells {
				for j := range cells[i] {
					if cells[i][j] != 0 {
						fieldCells[i][j] = -1.0
					} else {
						fieldCells[i][j] = 0.0
					}
				}
			}
			simPlayerMap := make([]*SimPlayer, 0)
			simPlayerMap = append(simPlayerMap, nil)
			newSimPlayer := SimPlayer{X: player.X, Y: player.Y, Direction: player.Direction, Speed: player.Speed, allVisitedCells: make(map[Coords]struct{})}
			simPlayerMap = append(simPlayerMap, &newSimPlayer)

		}
	}
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
func rolloutMove(status *Status, action Action) *Status {
	player := status.Players[status.You]
	turn := status.Turn
	if action == SpeedUp {
		if player.Speed != 10 {
			player.Speed++
		} else {
			player.Active = false
			return status
		}
	} else if action == SlowDown {
		if player.Speed != 1 {
			player.Speed--
		} else {
			player.Active = false
			return status
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
		inCells := player.Y >= 0 && player.Y < len(status.Cells) && player.X >= 0 && player.X < len(status.Cells[0])
		if !jump || i == 1 || i == player.Speed {
			if inCells && status.Cells[player.Y][player.X] == 0 {
				status.Cells[player.Y][player.X] = status.You
			} else {
				player.Active = false
				return status
			}
		}
	}
	return status
}

//searche for the longest paths a player could reach. Only simulates moves for one player!
func simulateRollouts(status *Status, limit int, ch chan [][]Action) [][]Action {
	longest := 0
	longestPaths := make([][]Action, 0)
	for j := 0; j < 20000; j++ {
		rolloutStatus := copyStatus(status)
		path := make([]Action, 0)
		for i := 0; i < limit; i++ {
			possibleMoves := Moves(rolloutStatus, rolloutStatus.Players[status.You], nil)
			if len(possibleMoves) == 0 {
				break
			}
			randomAction := possibleMoves[rand.Intn(len(possibleMoves))]
			rolloutStatus = rolloutMove(rolloutStatus, randomAction)
			if rolloutStatus.Players[status.You].Active == true {
				path = append(path, randomAction)

				rolloutStatus.Turn = rolloutStatus.Turn + 1
			} else {
				break
			}
		}
		if float32(len(path)) >= float32(longest)*0.8 {

			if longest >= len(path) {
				longestPaths = append(longestPaths, path)
			} else {

				longestPaths = make([][]Action, 0)
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

// SpekuClient is a client implementation that uses speculation to decide what to do next
type SpekuClient struct{}

// GetAction implements the Client interface
//TODO: use player information
func (c SpekuClient) GetAction(player Player, status *Status) Action {
	start := time.Now()
	if len(Moves(status, &player, nil)) == 1 {
		return Moves(status, &player, nil)[0]
	} else if len(Moves(status, &player, nil)) == 0 {
		return ChangeNothing
	}
	simChan := make(chan [][]Action, 1)
	go simulateRollouts(status, 10000, simChan)
	simDepth := 9
	fieldChan := make(chan [][]float64, simDepth)
	go simulateGame(status, fieldChan, simDepth)
	for i := 0; i < simDepth; i++ {
		//fmt.Println(<-fieldChan)
		fmt.Println("nix")
	}
	bestPaths := <-simChan
	fmt.Println(len(bestPaths))
	counterNothing := 0
	counterLeft := 0
	counterRight := 0
	counterUp := 0
	counterDown := 0
	for _, path := range bestPaths {
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
	valueNothing := float32(counterNothing) / float32(len(bestPaths))
	valueLeft := float32(counterLeft) / float32(len(bestPaths))
	valueRight := float32(counterRight) / float32(len(bestPaths))
	valueUp := float32(counterUp) / float32(len(bestPaths))
	valueDown := float32(counterDown) / float32(len(bestPaths))
	fmt.Println("Change Nothing: ", valueNothing)
	fmt.Println("Turn Left: ", valueLeft)
	fmt.Println("TurnRight: ", valueRight)
	fmt.Println("Speed Up: ", valueUp)
	fmt.Println("Slow Down: ", valueDown)
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println(elapsed)
	if valueNothing > valueLeft && valueNothing > valueRight && valueNothing > valueUp && valueNothing > valueDown {
		return ChangeNothing
	} else if valueLeft > valueRight && valueLeft > valueUp && valueLeft > valueDown {
		return TurnLeft
	} else if valueRight > valueUp && valueRight > valueDown {
		return TurnRight
	} else if valueUp > valueDown {
		return SpeedUp
	} else {
		return SlowDown
	}
}
