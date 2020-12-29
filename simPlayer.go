package main

type SimPlayer struct {
	player               *Player
	Probability          float64
	AllVisitedCells      map[Coords]struct{}
	LastMoveVisitedCells map[Coords]struct{}
}

// Copys a SimPlayer Object (Might be transfered to a util.go)
func (player *SimPlayer) copySimPlayer() *SimPlayer {
	var p SimPlayer
	p.player = player.player.copyPlayer()
	p.Probability = player.Probability
	p.LastMoveVisitedCells = make(map[Coords]struct{})
	p.AllVisitedCells = make(map[Coords]struct{})
	for k := range player.AllVisitedCells {
		p.AllVisitedCells[k] = struct{}{}
	}
	return &p
}
