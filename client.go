package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gorilla/websocket"
)

// Player contains information on a specific player. It is provided by the server,
type Player struct {
	X               int `json:"x"`
	Y               int `json:"y"`
	Direction       Direction
	StringDirection string `json:"direction"`
	Speed           int    `json:"speed"`
	Active          bool   `json:"active"`
	Name            string `json:"name"`
}

// Status contains all information on the current game status
type Status struct {
	Width    int             `json:"width"`
	Height   int             `json:"height"`
	Cells    [][]int         `json:"cells"`
	Players  map[int]*Player `json:"players"`
	You      int             `json:"you"`
	Running  bool            `json:"running"`
	Deadline string          `json:"deadline"`
	Turn     int
}

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

// Directions maps string direction representation to int representation
var Directions = map[string]Direction{
	"up":    Up,
	"down":  Down,
	"right": Right,
	"left":  Left,
}

// Direction contains the direction the player is facing
type Direction int

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

// Client represents a handler that decides what the specific player should do next
type Client interface {
	GetAction(player Player, status *Status) Action
}

func getClient() Client {
	var client Client
	switch os.Args[1] {
	case "minimax":
		client = MinimaxClient{}
		break
	case "left":
		client = LeftClient{}
		break
	case "right":
		client = RightClient{}
		break
	case "smart":
		client = SmartClient{}
		break
	case "mcts":
		client = MctsClient{}
		break
	case "speku":
		client = SpekuClient{}
		break
	default:
		log.Fatal("usage:", os.Args[0], "<client>")
	}
	return client
}

func setupLogging() *log.Logger {
	logger := log.New(os.Stdout, "[client] ", log.Lmsgprefix | log.LstdFlags)
	logger.Println("using client", os.Args[1])
	log.SetPrefix(fmt.Sprintf("[%s] ", os.Args[1]))
	log.SetFlags(log.Lmsgprefix | log.LstdFlags)
	return logger
}

func getUrl(logger *log.Logger) string {
	url := os.Getenv("URL")
	if url == "" {
		url = "ws://localhost:8080/spe_ed"
	}
	logger.Println("connecting to server", url)
	key := os.Getenv("KEY")
	if key != "" {
		url = fmt.Sprintf("%s?key=%s", url, key)
	}
	return url
}


func main() {
	if len(os.Args) <= 1 {
		log.Fatalln("usage:", os.Args[0], "<client>")
	}

	client := getClient()
	clientLogger := setupLogging()
	url := getUrl(clientLogger)

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		clientLogger.Fatalln("could not establish connection:", err)
	}
	defer c.Close()
	clientLogger.Println("connected to server")

	var status Status
	var input Input
	status.Turn = 1
	err = c.ReadJSON(&status)
	if err != nil {
		clientLogger.Fatalln("error on first ws read:", err)
	}

	clientLogger.Println("field dimensions:", status.Width, "x", status.Height)
	clientLogger.Println("number of players:", len(status.Players))
	for status.Running && status.Players[status.You].Active {
		clientLogger.Println("turn", status.Turn)
		clientLogger.Println("deadline", status.Deadline)
		for _, p := range status.Players {
			p.Direction = Directions[p.StringDirection]
		}
		action := client.GetAction(*status.Players[status.You], &status)
		status.Turn++

		input = Input{action}
		err = c.WriteJSON(&input)
		if err != nil {
			clientLogger.Fatalln("error on ws write:", err)
			break
		}
		err = c.ReadJSON(&status)
		if err != nil {
			clientLogger.Fatalln("error on ws read:", err)
			break
		}
		counter := 0
		for _, player := range status.Players {
			if player.Active {
				counter++
			}
		}
		if counter > 1 {
			clientLogger.Println("active players:", counter)
			if !status.Players[status.You].Active {
				clientLogger.Println("lost")
			}
		} else if counter == 1 {
			if status.Players[status.You].Active {
				clientLogger.Println("won")
			} else {
				clientLogger.Println("lost")
			}
		} else {
			clientLogger.Println("lost")
		}
	}
	clientLogger.Println("player inactive, disconnecting...")
}
