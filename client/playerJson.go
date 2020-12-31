package main

// JSONPlayer contains information on a specific player as returned by the server.
type JSONPlayer struct {
	X               int `json:"x"`
	Y               int `json:"y"`
	Direction       Direction `json:"-"`
	StringDirection string `json:"direction"`
	Speed           int    `json:"speed"`
	Active          bool   `json:"active"`
	Name            string `json:"name"`
}

// ConvertToPlayer converts a JSONPlayer to a Player
func (jsonPlayer *JSONPlayer) ConvertToPlayer() *Player {
	var player Player
	player.X = uint16(jsonPlayer.X)
	player.Y = uint16(jsonPlayer.Y)
	player.Speed = uint8(jsonPlayer.Speed)
	player.Direction = jsonPlayer.Direction
	return &player
}

// Copy copies a JSONPlayer
func (jsonPlayer *JSONPlayer) Copy() *JSONPlayer {
	player := *jsonPlayer
	return &player
}

