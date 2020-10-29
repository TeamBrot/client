package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Player struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Direction string `json:"direction"`
	Speed     int    `json:"speed"`
	Active    bool   `json:"active"`
	Name      string `json:"name"`
}

type Status struct {
	Width    int             `json:"width"`
	Height   int             `json:"height"`
	Cells    [][]int         `json:"cells"`
	Players  map[int]*Player `json:"players"`
	You      int             `json:"you"`
	Running  bool            `json:"running"`
	Deadline string          `json:"deadline"`
}

func main() {
	fmt.Println("test")
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/spe_ed", nil)
	if err != nil {
		fmt.Println("could not establish connection", err)
		return
	}
	defer c.Close()

	var status Status
	c.ReadJSON(&status)

	fmt.Println(status)
}
