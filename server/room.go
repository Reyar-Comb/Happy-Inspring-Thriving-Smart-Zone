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
	ID      int32
}

func NewRoom(id int32) *Room {
	return &Room{
		Players: make(map[int32]*Player),
		Engine:  NewGame(),
		ID:      id,
	}
}

func (r *Room) AddPlayer(player *Player) {
	if _, exists := r.Players[player.ID]; exists {
		return
	}
	r.Players[player.ID] = player
	fmt.Printf("Room: Added player %d\n", player.ID)
}

func (r *Room) RemovePlayer(playerID int32) {
	delete(r.Players, playerID)
	fmt.Printf("Room: Removed player %d\n", playerID)
}

func (r *Room) IsFull() bool {
	return len(r.Players) >= 2
}

func GetAnotherPlayer(room *Room, player *Player) *Player {
	for _, p := range room.Players {
		if p.ID != player.ID {
			return p
		}
	}
	return nil
}
