package server

type GameState struct {
	X int
	Y int
}

type Game struct {
}

func NewGame() *Game {
	return &Game{}
}

func (g *Game) UpdateState(player *Player, state *GameState) {
	if state.X < 0 {
		state.X = 0
	}
	if state.Y < 0 {
		state.Y = 0
	}
	player.GameState = state
}

func DecodeGameState(state *StatePacket) *GameState {
	return &GameState{
		X: int(state.X),
		Y: int(state.Y),
	}
}
