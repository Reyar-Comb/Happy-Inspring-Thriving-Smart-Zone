package server

import (
	"fmt"
	"math/rand/v2"
	"net"
)

type Player struct {
	ID       int32
	Addr     *net.UDPAddr
	Location *Location
	HP       int32
	Room     *Room
	State    PlayerState
}

func NewPlayer(addr *net.UDPAddr, id int32) *Player {
	player := &Player{
		ID:       id,
		Addr:     addr,
		Location: &Location{},
		HP:       100,
		State:    PlayerWaiting,
	}
	return player
}

type Room struct {
	Players  map[int32]*Player
	Engine   *Game
	ID       int32
	RoomCode int32
	State    RoomState
}

type RoomState int

const (
	RoomWaiting RoomState = iota
	RoomReady
	RoomPlaying
)

type PlayerState int

const (
	PlayerWaiting PlayerState = iota
	PlayerReady
	PlayerPlaying
)

func NewRoom(id int32) *Room {
	return &Room{
		Players:  make(map[int32]*Player),
		Engine:   NewGame(),
		ID:       id,
		State:    RoomWaiting,
		RoomCode: int32(rand.IntN(9000) + 1000), // 1000-9999
	}
}

func (r *Room) AddPlayer(player *Player) {
	if _, exists := r.Players[player.ID]; exists {
		return
	}
	r.Players[player.ID] = player
	player.Room = r

	fmt.Printf("Room: Added player %d\n", player.ID)
}

func (r *Room) RemovePlayer(playerID int32) {
	delete(r.Players, playerID)
	fmt.Printf("Room: Removed player %d\n", playerID)
}

func (r *Room) IsFull() bool {
	return len(r.Players) >= 2
}

func (r *Room) IsEmpty() bool {
	return len(r.Players) == 0
}

func GetAnotherPlayer(room *Room, player *Player) *Player {
	for _, p := range room.Players {
		if p.ID != player.ID {
			return p
		}
	}
	return nil
}

func (p *Player) GetOpponent() *Player {
	if p.Room == nil {
		return nil
	}
	return GetAnotherPlayer(p.Room, p)
}

func (p *Player) Ready() bool {
	p.State = PlayerReady
	for _, player := range p.Room.Players {
		if player.State != PlayerReady {
			return false
		}
	}
	return true
}
