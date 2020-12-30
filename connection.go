package main

import (
	"log"

	"github.com/gorilla/websocket"
)

// Input contains the action taken by the player and is sent as JSON to the server
type Input struct {
	Action string `json:"action"`
}

// Connection represents a connection to a game server
type Connection struct {
	Conn *websocket.Conn
	Turn int
}

// NewConnection creates a new connection with the specified configuration
func NewConnection(config Config) (Connection, error) {
	log.Println("trying to connect to", config.gameURL)
	c, _, err := websocket.DefaultDialer.Dial(config.GetWSURL(), nil)
	if err != nil {
		return Connection{nil, 0}, err
	}
	log.Println("connection established")
	return Connection{c, 0}, nil
}

// WriteAction writes the specified action to the game server
func (c *Connection) WriteAction(action Action) error {
	err := c.Conn.WriteJSON(&Input{action.String()})
	if err != nil {
		return err
	}
	return nil
}

// ReadStatus reads the status from the connection
func (c *Connection) ReadStatus() (*Status, *JSONStatus, error) {
	var jsonStatus JSONStatus
	err := c.Conn.ReadJSON(&jsonStatus)
	if err != nil {
		return nil, nil, err
	}
	for _, p := range jsonStatus.Players {
		p.Direction = Directions[p.StringDirection]
	}
	c.Turn++
	jsonStatus.Turn = c.Turn
	status := jsonStatus.ConvertToStatus()
	return status, &jsonStatus, nil
}

// Close closes the connection
func (c *Connection) Close() error {
	return c.Conn.Close()
}
