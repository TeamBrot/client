package main

import (
	"fmt"
	"math"
	"math/rand"
)

var c float64 = 1.25

// MCTSNode definition
type MCTSNode struct {
	status   *Status
	player   *Player
	parent   *MCTSNode
	children []*MCTSNode
	wins     float64
	plays    int
	UCTValue float64
	move     Action
}

func expansion(node *MCTSNode, player *Player) {
	for _, action := range moves(node.status, player) {
		newStatus := copyStatus(node.status)
		newPlayer := newStatus.Players[newStatus.You]
		doMove(newStatus, newPlayer, action)
		for _, currentPlayer := range newStatus.Players {
			if currentPlayer.Active && currentPlayer != newPlayer {
				possibleMoves := moves(newStatus, currentPlayer)
				if len(possibleMoves) > 0 {
					doMove(newStatus, currentPlayer, possibleMoves[rand.Intn(len(possibleMoves))])
				} else {
					currentPlayer.Active = false
				}
			}
		}
		node.children = append(node.children, &MCTSNode{parent: node, player: newPlayer, status: newStatus, move: action, children: make([]*MCTSNode, 0)})
	}

}

func copyPlayer(player *Player) *Player {
	var p Player
	p.Active = player.Active
	p.Direction = player.Direction
	p.Name = player.Name
	p.Speed = player.Speed
	p.X = player.X
	p.Y = player.Y
	return &p
}
func simulateMCTS(node *MCTSNode, player *Player, depth int) float64 {
	status := copyStatus(node.status)
	score := 0
	for i := 0; i < depth; i++ {
		if player.Active {
			for _, currentPlayer := range status.Players {
				if currentPlayer.Active {
					possibleMoves := moves(status, currentPlayer)
					if len(possibleMoves) > 0 {
						doMove(status, currentPlayer, possibleMoves[rand.Intn(len(possibleMoves))])
					} else {
						currentPlayer.Active = false
						score = score + 40
					}
				}
			}
			score++
		} else {
			score = score - 40
			break
		}
	}
	if score > 100 {
		return 1
	}
	return 0

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

func MCTS(node *MCTSNode, depth int) {
	selectedNode := selection(node)
	player := selectedNode.player
	expansion(selectedNode, player)
	for _, newNode := range selectedNode.children {
		var score float64
		for z := 0; z < 15; z++ {
			score = simulateMCTS(newNode, player, depth)
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

//MctsClient comment
type MctsClient struct{}

// GetAction implements the Client interface
//TODO: use player information
func (c MctsClient) GetAction(player Player, status *Status) Action {
	depth := 100
	firstNode := &MCTSNode{status: status, plays: 1, player: &player, children: make([]*MCTSNode, 0), parent: nil}
	fmt.Println(status.Deadline)
	for i := 0; i < 10000; i++ {
		MCTS(firstNode, depth)
	}
	bestNode := 0
	bestScore := -1000
	for i, childNode := range firstNode.children {
		fmt.Println(childNode.plays)
		fmt.Println(childNode.wins)
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
	printNode(firstNodeArray)
	return bestAction

}
