package main

// Action contains an action that the player could take
type Action uint8

const (
	// ChangeNothing goes straight
	ChangeNothing Action = iota
	// TurnLeft turns left
	TurnLeft
	// TurnRight turns right
	TurnRight
	// SpeedUp increases the player speed
	SpeedUp
	// SlowDown decreases the player speed
	SlowDown
)

// Actions contains all actions that could be taken
var Actions = []Action{ChangeNothing, SpeedUp, SlowDown, TurnLeft, TurnRight}

var actionNameMap = map[Action]string{
	ChangeNothing: "change_nothing",
	TurnLeft:      "turn_left",
	TurnRight:     "turn_right",
	SpeedUp:       "speed_up",
	SlowDown:      "slow_down",
}

func (action Action) String() string {
	return actionNameMap[action]
}
