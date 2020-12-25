package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

var c float64 = math.Sqrt2

// MCTSNode definition
type MCTSNode struct {
	status   *Status
	parent   *MCTSNode
	children []*MCTSNode
	wins     float64
	plays    int
	UCTValue float64
	move     Action
}

func expansion(node *MCTSNode) {
	player := node.status.Players[node.status.You]
	for _, action := range Moves(node.status, player, nil) {
		newStatus := copyStatus(node.status)
		//newPlayer := newStatus.Players[newStatus.You]
		nextActions := make(map[int]Action)
		nextActions[newStatus.You] = action
		for id, p := range newStatus.Players {
			if p.Active && p != newStatus.Players[newStatus.You] {
				actions, _ := bestActionsMinimax(id, newStatus.You, newStatus, 4, nil, nil)
				if len(actions) == 0 {
					nextActions[id] = ChangeNothing
				} else {
					nextActions[id] = actions[rand.Intn(len(actions))]
				}
			}
		}
		doMoves(newStatus, nextActions)
		node.children = append(node.children, &MCTSNode{parent: node, status: newStatus, move: action, children: make([]*MCTSNode, 0)})
	}
}

func simulateMCTS(node *MCTSNode, depth int) float64 {
	status := copyStatus(node.status)
	activeOtherPlayers := 0
	for id, p := range status.Players {
		if p.Active && id != status.You {
			activeOtherPlayers++
		}
	}
	surviveScore := 0.0
	for i := 0; i < depth; i++ {
		if status.Players[status.You].Active {
			nextActions := make(map[int]Action)
			for id, p := range status.Players {
				if p.Active {
					// Is checking for possible moves usefull?
					possibleMoves := Moves(status, p, nil)
					var randomAction Action
					if len(possibleMoves) > 0 {
						randomAction = possibleMoves[rand.Intn(len(possibleMoves))]
					} else {
						randomAction = Actions[rand.Intn(len(Actions))]
					}
					nextActions[id] = randomAction
				}
			}
			doMoves(status, nextActions)
			surviveScore++
		} else {
			surviveScore = surviveScore - float64(i/2)
			break
		}
	}
	stillActiveOtherPlayers := 0
	for id, p := range status.Players {
		if p.Active && id != status.You {
			stillActiveOtherPlayers++
		}
	}
	killScore := activeOtherPlayers - stillActiveOtherPlayers
	if killScore > 0 && surviveScore > float64(depth/4) || surviveScore > float64(depth/2) {
		return 10
	}
	return surviveScore / float64(depth)

	//	return float64(score) / float64(depth)
}
func backpropagation(node *MCTSNode, wins float64) {
	for node.parent != nil {
		node.wins = node.wins + wins
		node.plays++
		uct1 := float64(node.plays)
		uct2 := math.Log2(float64(node.parent.plays)) / math.Log2E
		uct3 := math.Sqrt(uct2 / uct1)
		node.UCTValue = (node.wins / float64(node.plays)) + c*uct3
		node = node.parent
	}
	node.plays++

}

func selection(node *MCTSNode) *MCTSNode {
	for len(node.children) != 0 {
		bestNode := 0
		bestScore := -1000.0
		for i, childNode := range node.children {
			if childNode.UCTValue > bestScore {
				bestScore = childNode.UCTValue
				bestNode = i
			}
		}
		node = node.children[bestNode]
	}
	return node
}

// MCTS function
func MCTS(node *MCTSNode, depth int) {
	selectedNode := selection(node)
	expansion(selectedNode)
	for _, newNode := range selectedNode.children {
		var score float64
		for z := 0; z < 1; z++ {
			score = simulateMCTS(newNode, depth)
			backpropagation(newNode, score)
		}
		//score = score / 15
	}
}

func printNode(nodes []*MCTSNode) {
	newNodes := make([]*MCTSNode, 0)
	for _, node := range nodes {
		fmt.Printf("M %s P %d W %.0f U %.0f \t", node.move, node.plays, node.wins, node.UCTValue)
		for _, childNode := range node.children {
			newNodes = append(newNodes, childNode)
		}
	}
	fmt.Print("\n")
	if len(newNodes) > 0 {
		printNode(newNodes)
	}
}

func doMoves(status *Status, moves map[int]Action) {
	occupiedCells := make([][]uint64, status.Height)
	for i := range occupiedCells {
		occupiedCells[i] = make([]uint64, status.Width)
	}
	jump := status.Turn%6 == 0
	var processedPlayers []int
	for id, p := range status.Players {
		processedPlayers = append(processedPlayers, id)
		action := moves[id]

		if action != "turn_left" && action != "turn_right" && action != "slow_down" && action != "speed_up" {
			action = "change_nothing"
		}
		if action == "speed_up" {
			if p.Speed != 10 {
				p.Speed++
			}
		} else if action == "slow_down" {
			if p.Speed != 1 {
				p.Speed--
			}
		} else if action == "turn_left" {
			switch p.Direction {
			case Left:
				p.Direction = Down
				break
			case Down:
				p.Direction = Right
				break
			case Right:
				p.Direction = Up
				break
			case Up:
				p.Direction = Left
				break
			}
		} else if action == "turn_right" {
			switch p.Direction {
			case Left:
				p.Direction = Up
				break
			case Down:
				p.Direction = Left
				break
			case Right:
				p.Direction = Down
				break
			case Up:
				p.Direction = Right
				break
			}
		}

		for i := 1; i <= p.Speed; i++ {
			if p.Direction == Up {
				p.Y--
			} else if p.Direction == Down {
				p.Y++
			} else if p.Direction == Right {
				p.X++
			} else if p.Direction == Left {
				p.X--
			}

			if p.X >= status.Width || p.Y >= status.Height || p.X < 0 || p.Y < 0 {
				p.Active = false
				break
			}

			if !jump || i == 1 || i == p.Speed {
				// If the cell is non-empty set error bit and do not write to the cell
				if status.Cells[p.Y][p.X] != 0 {
					occupiedCells[p.Y][p.X] |= 1
				} else {
					status.Cells[p.Y][p.X] = id
				}
				occupiedCells[p.Y][p.X] |= (1 << id)
			}
		}

	}

	for y := range occupiedCells {
		for x := range occupiedCells[y] {
			if occupiedCells[y][x] != 0 {
				for _, playerID := range processedPlayers {
					// If the error bit is set and the (playerID)th bit as well, kill player and set cell to -1
					if occupiedCells[y][x]&1 != 0 && occupiedCells[y][x]&(1<<playerID) != 0 {
						//log.Print("Player ", playerID, " moved to field: ", y, " ", x)
						status.Cells[y][x] = -1
						status.Players[playerID].Active = false
					}
				}
			}
		}
	}

}

//MctsClient comment
type MctsClient struct{}

// GetAction implements the Client interface
//TODO: use player information
func (c MctsClient) GetAction(player Player, status *Status, timingChannel <-chan time.Time) Action {
	depth := 70
	firstNode := &MCTSNode{status: status, plays: 1, children: make([]*MCTSNode, 0), parent: nil}
	for i := 0; i < 2000; i++ {
		MCTS(firstNode, depth)
	}
	bestNode := 0
	bestScore := -1000
	for i, childNode := range firstNode.children {
		fmt.Printf("%s \t %d \t %.2f \n", childNode.move, childNode.plays, childNode.wins)
		if childNode.plays > bestScore {
			bestScore = childNode.plays
			bestNode = i
		}
	}
	var bestAction Action
	if len(firstNode.children) != 0 {
		bestAction = firstNode.children[bestNode].move
	} else {
		bestAction = ChangeNothing
	}
	firstNodeArray := make([]*MCTSNode, 0)
	firstNodeArray = append(firstNodeArray, firstNode)
	//printNode(firstNodeArray)
	return bestAction

}
