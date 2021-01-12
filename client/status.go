package main

import (
	"errors"
	"math"
)

// Status contains all information relevant to the clients
type Status struct {
	Width   uint16
	Height  uint16
	Cells   [][]bool
	Players map[uint8]*Player
	You     uint8
	Turn    uint16
}

// Copy copies a status
func (status *Status) Copy() *Status {
	var s Status
	s.Width = status.Width
	s.Height = status.Height
	s.Turn = status.Turn
	s.You = status.You
	s.Cells = make([][]bool, s.Height)
	for i := range s.Cells {
		s.Cells[i] = make([]bool, status.Width)
		for j := range s.Cells[i] {
			s.Cells[i][j] = status.Cells[i][j]
		}
	}
	s.Players = make(map[uint8]*Player)
	for id, player := range status.Players {
		s.Players[id] = player.Copy()
	}
	return &s
}

// FindClosestPlayerTo returns the player that is closest to the specified
func (status *Status) FindClosestPlayerTo(originPlayer uint8) (uint8, error) {
	ourPlayer := status.Players[originPlayer]
	var nearestPlayer uint8
	nearestPlayerDistance := math.Inf(0)
	for playerID, player := range status.Players {
		distance := ourPlayer.DistanceTo(player)
		if playerID != originPlayer && distance < nearestPlayerDistance {
			nearestPlayer = playerID
			nearestPlayerDistance = distance
		}
	}
	if nearestPlayer == 0 {
		return 0, errors.New("no non-dead player found")
	}
	return nearestPlayer, nil
}
