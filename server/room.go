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

func (r *Room) GetReadyStatus() (int32, byte, int32, byte) {
	if len(r.Players) == 0 {
		return -1, 0x00, -1, 0x00
	} else if len(r.Players) == 1 {
		for _, p := range r.Players {
			return p.ID, p.Ready(), -1, 0x00
		}
	} else {
		i := 0
		var pid1, pid2 int32
		var rdy1, rdy2 byte
		for _, p := range r.Players {
			if i == 0 {
				pid1 = p.ID
				rdy1 = p.Ready()
			} else {
				pid2 = p.ID
				rdy2 = p.Ready()
			}
			i++
		}
		return pid1, rdy1, pid2, rdy2
	}
	return -1, 0x00, -1, 0x00
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

func (p *Player) SetReady() {
	p.State = PlayerReady
}

func (p *Player) SetUnready() {
	p.State = PlayerWaiting
}

func (p *Player) Ready() byte {
	if p.State == PlayerReady {
		return 0x01
	}
	return 0x00
}
