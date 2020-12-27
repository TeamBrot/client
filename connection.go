package main

import (
	"log"

	"github.com/gorilla/websocket"
)

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
	status := JSONStatus.ConvertToStatus(JSONstatus)
	return status, &JSONstatus, nil
}

// Close closes the connection
func (c *Connection) Close() error {
	return c.Conn.Close()
}
