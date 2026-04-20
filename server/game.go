package server

type Location struct {
	X int
	Y int
}

type Game struct {
}

func NewGame() *Game {
	return &Game{}
}

func (g *Game) UpdateLocation(player *Player, location *Location) {
	if location.X < 0 {
		location.X = 0
	}
	if location.Y < 0 {
		location.Y = 0
	}
	player.Location = location
}

func (g *Game) UpdateHp(player *Player, value int32) bool {
	prevHP := player.HP
	player.HP += value
	if player.HP < 0 && prevHP > 0 {
		return false
	}
	if player.HP > 100 {
		player.HP = 100
	}
	return true
}

func (g *Game) GameOver(winner *Player) {
	if winner == nil {
		return
	}
	room := winner.Room
	if room == nil {
		return
	}
}

func DecodeLocation(locPacket *LocationPacket) *Location {
	return &Location{
		X: int(locPacket.X),
		Y: int(locPacket.Y),
	}
}
