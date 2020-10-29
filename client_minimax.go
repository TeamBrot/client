package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Player struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Direction string `json:"direction"`
	Speed     int    `json:"speed"`
	Active    bool   `json:"active"`
	Name      string `json:"name"`
}

type Status struct {
	Width    int             `json:"width"`
	Height   int             `json:"height"`
	Cells    [][]int         `json:"cells"`
	Players  map[int]*Player `json:"players"`
	You      int             `json:"you"`
	Running  bool            `json:"running"`
	Deadline string          `json:"deadline"`
}

type Input struct {
	Action string `json:"action"`
}

func checkCell(status *Status, current bool, y int, x int) bool {
	if x >= status.Width || y >= status.Height || x < 0 || y < 0 {
		return false
	}
	return status.Cells[y][x] == 0 && current
}

/* add jumping */
func moves(status *Status, player *Player) []string {
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

	possibleMoves := make([]string, 0)

	if speedDown && player.Speed != 1 {
		possibleMoves = append(possibleMoves, "slow_down")
	}
	if changeNothing {
		possibleMoves = append(possibleMoves, "change_nothing")
	}
	if speedUp && player.Speed != 10 {
		possibleMoves = append(possibleMoves, "speed_up")
	}
	if turnLeft {
		possibleMoves = append(possibleMoves, "turn_left")
	}
	if turnRight {
		possibleMoves = append(possibleMoves, "turn_right")
	}
	return possibleMoves
}

var ACTIONS = []string{"change_nothing", "speed_up", "slow_down", "turn_left", "turn_right"}

func simulate(player Player, status *Status, action string) int {
	if action == "speed_up" {
		if player.Speed != 10 {
			player.Speed++
		}
	} else if action == "slow_down" {
		if player.Speed != 1 {
			player.Speed--
		}
	} else if action == "turn_left" {
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
	} else if action == "turn_right" {
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

func main() {
	fmt.Println("test")
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/spe_ed", nil)
	if err != nil {
		fmt.Println("could not establish connection", err)
		return
	}
	defer c.Close()

	var status Status
	var input Input
	err = c.ReadJSON(&status)
	if err != nil {
		return
	}

	for status.Players[status.You].Active {
		bestAction := ""
		bestScore := -1
		for _, action := range moves(&status, status.Players[status.You]) {
			score := simulate(*status.Players[status.You], &status, action)
			if score > bestScore {
				bestAction = action
				bestScore = score
			}
		}

		input = Input{bestAction}
		err = c.WriteJSON(&input)
		if err != nil {
			break
		}
		err = c.ReadJSON(&status)
		if err != nil {
			break
		}
	}
}
