package server

import (
	"fmt"
	"net"

	"github.com/Reyar-Comb/HITPlane/config"
)

type Server struct {
	Addr   string
	Conn   *net.UDPConn
	Rooms  []*Room
	Sender *Sender
}

var id int32 = 1001

func Start(s *Server) error {
	udpAddr, err := net.ResolveUDPAddr("udp", config.GlobalConfig.Port)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	s.Conn = conn
	s.Sender = NewSender(conn)
	defer conn.Close()

	fmt.Println("Server: Listening on", s.Addr)

	buffer := make([]byte, 1024)

	for {
		n, clientAddr, err := s.Conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Server: Error reading from UDP:", err)
			continue
		}

		packet := buffer[:n]

		s.handlePacket(packet, clientAddr)
	}
}

func (s *Server) AddRoom(room *Room) {
	s.Rooms = append(s.Rooms, room)
	fmt.Printf("Server: Added new room, total rooms: %d\n", len(s.Rooms))
}

func (s *Server) handlePacket(packet []byte, clientAddr *net.UDPAddr) {
	if len(packet) == 0 {
		return
	}

	opCode := packet[0]

	switch opCode {
	case OpStateUpdate:
		state, err := DecodeStatePacket(packet)
		if err != nil {
			fmt.Printf("Server: Error decoding state packet from %s: %v\n", clientAddr, err)
			return
		}

		if len(s.Rooms) > 0 {
			room := s.Rooms[0]

			player, exsists := room.Players[state.PlayerID]
			if !exsists {
				return
			}

			room.Engine.UpdateState(player, DecodeGameState(state))

			s.Sender.PlayerBroadcast(
				room,
				state.PlayerID,
				EncodeStatePacket(
					&StatePacket{
						PlayerID: state.PlayerID,
						X:        int32(player.GameState.X),
						Y:        int32(player.GameState.Y),
					},
				),
			)
		}

	case OpJoin:
		if len(s.Rooms) == 0 {
			room := NewRoom()
			s.AddRoom(room)
		}

		room := s.Rooms[0]

		playerID := id
		id++

		player := NewPlayer(clientAddr, playerID)
		room.AddPlayer(player)

		s.Sender.SendTo(
			player.Addr,
			EncodeAcceptPacket(
				&AcceptPacket{PlayerID: player.ID},
			))

	}
}
