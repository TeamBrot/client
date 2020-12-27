package main

import (
	"errors"
	"time"
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

// JSONStatus contains all information on the current game status
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
		s.Players[id] = player.copyPlayer()
	}
	return &s
}

func (status *Status) FindClosestPlayerTo(originPlayer uint8) (uint8, error) {
	ourPlayer := status.Players[originPlayer]
	var nearestPlayer uint8
	nearestPlayerDistance := 0.0
	for playerID, player := range status.Players {
		distance := distanceToPlayer(player, ourPlayer) //math.Sqrt(math.Pow(float64(player.X-ourPlayer.X), 2) + math.Pow(float64(player.Y-ourPlayer.Y), 2))
		if playerID != originPlayer && (nearestPlayer == 0 || distance < nearestPlayerDistance) {
			nearestPlayer = playerID
			nearestPlayerDistance = distance
		}
	}
	if nearestPlayer == 0 {
		return 0, errors.New("no non-dead player found")
	}
	return nearestPlayer, nil
}

func (js JSONStatus) ConvertToStatus() *Status {
	var status Status
	status.Height = uint16(js.Height)
	status.Turn = uint16(js.Turn)
	status.Width = uint16(js.Width)
	status.You = uint8(js.You)
	status.Players = make(map[uint8]*Player, 0)
	for z, JSONPlayer := range js.Players {
		status.Players[uint8(z)] = JSONPlayer.ConvertToPlayer()
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

