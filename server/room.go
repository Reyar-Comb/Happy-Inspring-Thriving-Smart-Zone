package server

import (
	"fmt"
	"net"
)

type Player struct {
	ID       int32
	Addr     *net.UDPAddr
	Location *Location
	HP       int32
}

func NewPlayer(addr *net.UDPAddr, id int32) *Player {
	player := &Player{
		ID:       id,
		Addr:     addr,
		Location: &Location{},
		HP:       100,
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
