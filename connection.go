package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Status struct {
	Width   uint16
	Height  uint16
	Cells   [][]bool
	Players map[uint8]*Player
	You     uint8
	Turn    uint16
}

// Status contains all information on the current game status
type JSONStatus struct {
	Width    int                 `json:"width"`
	Height   int                 `json:"height"`
	Cells    [][]int             `json:"cells"`
	Players  map[int]*JSONPlayer `json:"players"`
	You      int                 `json:"you"`
	Running  bool                `json:"running"`
	Deadline time.Time           `json:"deadline"`
	Turn     int
}

type Player struct {
	X         uint16
	Y         uint16
	Direction Direction
	Speed     uint8
}

// Player contains information on a specific player. It is provided by the server,
type JSONPlayer struct {
	X               int `json:"x"`
	Y               int `json:"y"`
	Direction       Direction
	StringDirection string `json:"direction"`
	Speed           int    `json:"speed"`
	Active          bool   `json:"active"`
	Name            string `json:"name"`
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

// Directions maps string direction representation to int representation
var Directions = map[string]Direction{
	"up":    Up,
	"down":  Down,
	"right": Right,
	"left":  Left,
}

// Direction contains the direction the player is facing
type Direction uint8

// turning left is equivalent to +1(mod 4) and turning right to (+3)(mod 4)
const (
	// Up makes the player face up
	Up Direction = 0
	// Left makes the player face left
	Left = 1
	// Down makes the player face down
	Down = 2
	// Right makes the player face right
	Right = 3
)

// Connection represents a connection to a game server
type Connection struct {
	Conn *websocket.Conn
	Turn int
}

// NewConnection creates a new connection with the specified configuration
func NewConnection(config Config) (Connection, error) {
	log.Println("trying to connect to", config.GameURL)
	c, _, err := websocket.DefaultDialer.Dial(config.GetWSURL(), nil)
	if err != nil {
		return Connection{nil, 0}, err
	}
	log.Println("connection established")
	return Connection{c, 0}, nil
}

func (JSONPlayer *JSONPlayer) ToPlayer() *Player {
	var player Player
	player.X = uint16(JSONPlayer.X)
	player.Y = uint16(JSONPlayer.Y)
	player.Speed = uint8(JSONPlayer.Speed)
	player.Direction = JSONPlayer.Direction
	return &player
}

//
func (js JSONStatus) ToStatus() *Status {
	var status Status
	status.Height = uint16(js.Height)
	status.Turn = uint16(js.Turn)
	status.Width = uint16(js.Width)
	status.You = uint8(js.You)
	status.Players = make(map[uint8]*Player, 0)
	for z, JSONPlayer := range js.Players {
		if JSONPlayer.Active {
			status.Players[uint8(z)] = JSONPlayer.ToPlayer()
		}
	}
	status.Cells = make([][]bool, status.Height)
	for y := range status.Cells {
		status.Cells[y] = make([]bool, status.Width)
	}
	for y := range js.Cells {
		for x := range js.Cells[0] {
			if js.Cells[y][x] != 0 {
				status.Cells[y][x] = true
			}
		}
	}
	return &status
}

// WriteAction writes the specified action to the game server
func (c *Connection) WriteAction(action Action) error {
	err := c.Conn.WriteJSON(&Input{action})
	if err != nil {
		return err
	}
	return nil
}

// ReadStatus reads the status from the connection
func (c *Connection) ReadStatus() (*Status, *JSONStatus, error) {
	var JSONstatus JSONStatus
	err := c.Conn.ReadJSON(&JSONstatus)
	if err != nil {
		return nil, nil, err
	}
	for _, p := range JSONstatus.Players {
		p.Direction = Directions[p.StringDirection]
	}
	c.Turn++
	JSONstatus.Turn = c.Turn
	status := JSONStatus.ToStatus(JSONstatus)
	return status, &JSONstatus, nil
}

// Close closes the connection
func (c *Connection) Close() error {
	return c.Conn.Close()
}
