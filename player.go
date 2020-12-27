package main

// Player contains information on a specific player used by the API
type Player struct {
	X         uint16
	Y         uint16
	Direction Direction
	Speed     uint8
}

// JSONPlayer contains information on a specific player as returned by the server.
type JSONPlayer struct {
	X               int `json:"x"`
	Y               int `json:"y"`
	Direction       Direction
	StringDirection string `json:"direction"`
	Speed           int    `json:"speed"`
	Active          bool   `json:"active"`
	Name            string `json:"name"`
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

// Directions maps string direction representation to int representation
var Directions = map[string]Direction{
	"up":    Up,
	"down":  Down,
	"right": Right,
	"left":  Left,
}


// Convert a JSONPlayer to a Player
func (JSONPlayer *JSONPlayer) ConvertToPlayer() *Player {
	var player Player
	player.X = uint16(JSONPlayer.X)
	player.Y = uint16(JSONPlayer.Y)
	player.Speed = uint8(JSONPlayer.Speed)
	player.Direction = JSONPlayer.Direction
	return &player
}

