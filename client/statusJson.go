package main

import (
	"time"
)

// JSONStatus contains all information on the current game status
type JSONStatus struct {
	Width    int                 `json:"width"`
	Height   int                 `json:"height"`
	Cells    [][]int             `json:"cells"`
	Players  map[int]*JSONPlayer `json:"players"`
	You      int                 `json:"you"`
	Running  bool                `json:"running"`
	Deadline time.Time           `json:"deadline"`
	Turn     int                 `json:"turn"`
}

func (jsonStatus *JSONStatus) Copy() *JSONStatus {
	newStatus := *jsonStatus
	newStatus.Cells = make([][]int, newStatus.Height)
	for i := range newStatus.Cells {
		newStatus.Cells[i] = make([]int, jsonStatus.Width)
		for j := range newStatus.Cells[i] {
			newStatus.Cells[i][j] = jsonStatus.Cells[i][j]
		}
	}
	newStatus.Players = make(map[int]*JSONPlayer)
	for id, player := range jsonStatus.Players {
		newStatus.Players[id] = player.Copy()
	}
	return &newStatus

}

func (js JSONStatus) ConvertToStatus() *Status {
	var status Status
	status.Height = uint16(js.Height)
	status.Turn = uint16(js.Turn)
	status.Width = uint16(js.Width)
	status.You = uint8(js.You)
	status.Players = make(map[uint8]*Player, 0)
	for z, jsonPlayer := range js.Players {
		if jsonPlayer.Active {
			status.Players[uint8(z)] = jsonPlayer.ConvertToPlayer()
		}
	}
	status.Cells = make([][]bool, status.Height)
	for y := range status.Cells {
		status.Cells[y] = make([]bool, status.Width)
	}
	for y := range js.Cells {
		for x := range js.Cells[0] {
			status.Cells[y][x] = js.Cells[y][x] != 0
		}
	}
	return &status
}

