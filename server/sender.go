package server

import (
	"fmt"
	"net"
)

type Sender struct {
	Conn *net.UDPConn
}

func NewSender(conn *net.UDPConn) *Sender {
	return &Sender{Conn: conn}
}

func (s *Sender) SendTo(addr *net.UDPAddr, data []byte) {
	_, err := s.Conn.WriteToUDP(data, addr)
	if err != nil {
		fmt.Printf("Sender: Error sending to %s: %v\n", addr, err)
	}
}

func (s *Sender) PlayerBroadcast(room *Room, senderID int32, data []byte) {
	for id, player := range room.Players {
		if id != senderID {
			s.SendTo(player.Addr, data)
		}
	}
}

func (s *Sender) RoomBroadcast(room *Room, data []byte) {
	for _, player := range room.Players {
		s.SendTo(player.Addr, data)
	}
}
