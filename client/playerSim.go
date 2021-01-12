package main

// SimPlayer adds a new array of visited cells
type SimPlayer struct {
	player               *Player
	Probability          float64
	AllVisitedCells      map[Coords]struct{}
	LastMoveVisitedCells map[Coords]struct{}
}

// Copy copies a SimPlayer object
func (player *SimPlayer) Copy() *SimPlayer {
	var p SimPlayer
	p.player = player.player.Copy()
	p.Probability = player.Probability
	p.LastMoveVisitedCells = make(map[Coords]struct{})
	p.AllVisitedCells = make(map[Coords]struct{})
	for k := range player.AllVisitedCells {
		p.AllVisitedCells[k] = struct{}{}
	}
	return &p
}

// SimPlayerFromPlayer is the constructor for a simPlayer. A simPlayer can only be constructed with a player
func SimPlayerFromPlayer(player *Player, probability float64) *SimPlayer {
	var simPlayer SimPlayer
	simPlayer.player = player.Copy()
	simPlayer.AllVisitedCells = make(map[Coords]struct{}, 0)
	simPlayer.LastMoveVisitedCells = make(map[Coords]struct{}, 0)
	simPlayer.Probability = probability
	return &simPlayer
}
