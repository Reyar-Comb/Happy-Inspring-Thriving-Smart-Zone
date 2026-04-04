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
	case OpLocationUpdate:
		state, err := DecodeLocationPacket(packet)
		if err != nil {
			fmt.Printf("Server: Error decoding location packet from %s: %v\n", clientAddr, err)
			return
		}

		if len(s.Rooms) > 0 {
			room := s.Rooms[0]

			player, exsists := room.Players[state.PlayerID]
			if !exsists {
				return
			}

			room.Engine.UpdateLocation(player, DecodeLocation(state))

			s.Sender.PlayerBroadcast(
				room,
				state.PlayerID,
				EncodeLocationPacket(
					&LocationPacket{
						PlayerID: state.PlayerID,
						X:        int32(player.Location.X),
						Y:        int32(player.Location.Y),
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

	case OpHit:
		hitPacket, err := DecodeHitPacket(packet)
		if err != nil {
			fmt.Printf("Server: Error decoding hit packet from %s: %v\n", clientAddr, err)
			return
		}

		if len(s.Rooms) > 0 {
			room := s.Rooms[0]

			player, exists := room.Players[hitPacket.PlayerID]
			if !exists {
				return
			}
			target := GetAnotherPlayer(room, player)
			if target == nil {
				return
			}
			room.Engine.UpdateHp(target, -hitPacket.Damage)

			s.Sender.RoomBroadcast(
				room,
				EncodeHpPacket(
					&HpPacket{
						PlayerID: target.ID,
						Hp:       target.HP,
					},
				),
			)
		}

	case OpShoot:
		shootPacket, err := DecodeShootPacket(packet)
		if err != nil {
			fmt.Printf("Server: Error decoding shoot packet from %s: %v\n", clientAddr, err)
			return
		}

		if len(s.Rooms) > 0 {
			room := s.Rooms[0]

			_, exists := room.Players[shootPacket.PlayerID]
			if !exists {
				return
			}

			s.Sender.PlayerBroadcast(
				room,
				shootPacket.PlayerID,
				EncodeShootPacket(shootPacket),
			)
		}

		fmt.Println("Server: Shoot")
	}
}
