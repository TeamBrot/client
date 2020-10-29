package main

import (
	"fmt"
	"os"

	"github.com/gorilla/websocket"
)

// Player contains information on a specific player. It is provided by the server,
type Player struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Direction string `json:"direction"`
	Speed     int    `json:"speed"`
	Active    bool   `json:"active"`
	Name      string `json:"name"`
}

// Status contains all information on the current game status
type Status struct {
	Width    int             `json:"width"`
	Height   int             `json:"height"`
	Cells    [][]int         `json:"cells"`
	Players  map[int]*Player `json:"players"`
	You      int             `json:"you"`
	Running  bool            `json:"running"`
	Deadline string          `json:"deadline"`
}

// Input contains the action taken by the player and is sent as JSON to the server
type Input struct {
	Action Action `json:"action"`
}

// Action contains an action that the player could take
type Action string

const (
	// ChangeNothing goes straight
	ChangeNothing Action = "change_nothing"
	// TurnLeft turns left
	TurnLeft = "turn_left"
	// TurnRight turns right
	TurnRight = "turn_right"
	// SpeedUp increases the player speed
	SpeedUp = "speed_up"
	// SlowDown decreases the player speed
	SlowDown = "slow_down"
)

// Actions contains all actions that could be taken
var Actions = []Action{ChangeNothing, SpeedUp, SlowDown, TurnLeft, TurnRight}

// Client represents a handler that decides what the specific player should do next
type Client interface {
	GetAction(player Player, status *Status) Action
}

func main() {
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

	var client Client
	switch os.Args[1] {
	case "minimax":
		client = MinimaxClient{}
		break
	}

	for status.Players[status.You].Active {
		action := client.GetAction(*status.Players[status.You], &status)

		input = Input{action}
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
