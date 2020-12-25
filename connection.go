package main

import (
	"time"

	"github.com/gorilla/websocket"
)

// Status contains all information on the current game status
type Status struct {
	Width    int             `json:"width"`
	Height   int             `json:"height"`
	Cells    [][]int         `json:"cells"`
	Players  map[int]*Player `json:"players"`
	You      int             `json:"you"`
	Running  bool            `json:"running"`
	Deadline time.Time       `json:"deadline"`
	Turn     int
}

// Player contains information on a specific player. It is provided by the server,
type Player struct {
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
type Direction int

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

type Connection struct {
	conn *websocket.Conn
	turn int
}

func NewConnection(config Config) (Connection, error) {
	c, _, err := websocket.DefaultDialer.Dial(config.GetWSURL(), nil)
	if err != nil {
		return Connection{nil, 0}, err
	}
	return Connection{c, 0}, nil
}

func (c *Connection) WriteAction(action Action) error {
	err := c.conn.WriteJSON(&Input{action})
	if err != nil {
		return err
	}
	return nil
}

func (c *Connection) ReadStatus() (*Status, error) {
	var status Status
	err := c.conn.ReadJSON(&status)
	if err != nil {
		return nil, err
	}
	for _, p := range status.Players {
		p.Direction = Directions[p.StringDirection]
	}
	c.turn++
	status.Turn = c.turn
	return &status, nil
}

func (c *Connection) Close() error {
	return c.conn.Close()
}
