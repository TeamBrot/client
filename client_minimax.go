package main

func checkCell(status *Status, current bool, y int, x int) bool {
	if x >= status.Width || y >= status.Height || x < 0 || y < 0 {
		return false
	}
	return status.Cells[y][x] == 0 && current
}

/* TODO add jumping */
func moves(status *Status, player *Player) []Action {
	changeNothing := true
	turnRight := true
	turnLeft := true
	speedDown := true
	for i := 1; i <= player.Speed; i++ {
		if player.Direction == "right" {
			turnRight = checkCell(status, turnRight, player.Y+i, player.X)
			changeNothing = checkCell(status, changeNothing, player.Y, player.X+i)
			if i != player.Speed {
				speedDown = changeNothing
			}
			turnLeft = checkCell(status, turnLeft, player.Y-i, player.X)
		} else if player.Direction == "up" {
			turnRight = checkCell(status, turnRight, player.Y, player.X+i)
			changeNothing = checkCell(status, changeNothing, player.Y-i, player.X)
			turnLeft = checkCell(status, turnLeft, player.Y, player.X-i)
		} else if player.Direction == "left" {
			turnRight = checkCell(status, turnRight, player.Y-i, player.X)
			changeNothing = checkCell(status, changeNothing, player.Y, player.X-i)
			turnLeft = checkCell(status, turnLeft, player.Y+i, player.X)
		} else if player.Direction == "down" {
			turnRight = checkCell(status, turnRight, player.Y, player.X-i)
			changeNothing = checkCell(status, changeNothing, player.Y+i, player.X)
			turnLeft = checkCell(status, turnLeft, player.Y, player.X+i)
		}
	}
	speedUp := changeNothing
	if player.Direction == "right" {
		speedUp = checkCell(status, speedUp, player.Y, player.X+player.Speed+1)
	} else if player.Direction == "up" {
		speedUp = checkCell(status, speedUp, player.Y-player.Speed-1, player.X)
	} else if player.Direction == "left" {
		speedUp = checkCell(status, speedUp, player.Y, player.X-player.Speed-1)
	} else if player.Direction == "down" {
		speedUp = checkCell(status, speedUp, player.Y+player.Speed+1, player.X)
	}

	possibleMoves := make([]Action, 0)

	if speedDown && player.Speed != 1 {
		possibleMoves = append(possibleMoves, SlowDown)
	}
	if changeNothing {
		possibleMoves = append(possibleMoves, ChangeNothing)
	}
	if speedUp && player.Speed != 10 {
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

func simulate(player Player, status *Status, action Action) int {
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
		case "left":
			player.Direction = "down"
			break
		case "down":
			player.Direction = "right"
			break
		case "right":
			player.Direction = "up"
			break
		case "up":
			player.Direction = "left"
			break
		}
	} else if action == TurnRight {
		switch player.Direction {
		case "left":
			player.Direction = "up"
			break
		case "down":
			player.Direction = "left"
			break
		case "right":
			player.Direction = "down"
			break
		case "up":
			player.Direction = "right"
			break
		}
	}

	for i := 1; i <= player.Speed; i++ {
		if player.Direction == "up" {
			player.Y--
		} else if player.Direction == "down" {
			player.Y++
		} else if player.Direction == "right" {
			player.X++
		} else if player.Direction == "left" {
			player.X--
		}

		jump := false
		if !jump || i == 1 || i == player.Speed {
			if status.Cells[player.Y][player.X] == 0 {
				status.Cells[player.Y][player.X] = status.You
				// defer func() { status.Cells[player.Y][player.X] = 0 }()
			} else {
				panic("the field should always be 0")
			}
		}
	}
	score := len(moves(status, &player))
	for i := 1; i <= player.Speed; i++ {
		status.Cells[player.Y][player.X] = 0
		if player.Direction == "up" {
			player.Y++
		} else if player.Direction == "down" {
			player.Y--
		} else if player.Direction == "right" {
			player.X--
		} else if player.Direction == "left" {
			player.X++
		}
	}

	return score
}

// MinimaxClient is a client implementation that uses Minimax to decide what to do next
type MinimaxClient struct{}

// GetAction implements the Client interface
func (c MinimaxClient) GetAction(player Player, status *Status) Action {
	var bestAction Action
	bestScore := -1
	for _, action := range moves(status, status.Players[status.You]) {
		score := simulate(*status.Players[status.You], status, action)
		if score > bestScore {
			bestAction = action
			bestScore = score
		}
	}
	return bestAction
}
