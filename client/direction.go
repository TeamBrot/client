package main

// Direction contains the direction the player is facing
type Direction uint8

// turning left is equivalent to +1(mod 4) and turning right to (+3)(mod 4)
const (
	// Up makes the player face up
	Up Direction = iota
	// Left makes the player face left
	Left
	// Down makes the player face down
	Down
	// Right makes the player face right
	Right
)

// Directions maps string direction representation to int representation
var Directions = map[string]Direction{
	"up":    Up,
	"down":  Down,
	"right": Right,
	"left":  Left,
}
