package server

import (
	"fmt"
	"net"
)

type Player struct {
	ID        int32
	Addr      *net.UDPAddr
	GameState *GameState
}

func NewPlayer(addr *net.UDPAddr, id int32) *Player {
	player := &Player{
		ID:        id,
		Addr:      addr,
		GameState: &GameState{},
	}
	return player
}

type Room struct {
	Players map[int32]*Player
	Engine  *Game
}

func NewRoom() *Room {
	return &Room{
		Players: make(map[int32]*Player),
		Engine:  NewGame(),
	}
}

func (r *Room) AddPlayer(player *Player) {
	if _, exists := r.Players[player.ID]; exists {
		return
	}
	r.Players[player.ID] = player
	fmt.Printf("Room: Added player %d\n", player.ID)
}
