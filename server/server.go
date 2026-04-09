package server

import (
	"fmt"
	"net"

	"github.com/Reyar-Comb/HITPlane/config"
)

type Server struct {
	Addr           string
	Conn           *net.UDPConn
	Rooms          map[int32]*Room
	AvailableRooms map[int32]*Room
	Sender         *Sender
}

var playerId int32 = 1001
var roomId int32 = 1

var PlayerRoomID = make(map[int32]int32)

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
	s.Rooms[room.ID] = room
	fmt.Printf("Server: Added new room, total rooms: ")
	for id := range s.Rooms {
		fmt.Printf("%d ", id)
	}
	fmt.Printf("\n")
}

func (s *Server) RemoveRoom(roomID int32) {
	delete(s.Rooms, roomID)
	fmt.Printf("Server: Removed room %d, total rooms: %d\n", roomID, len(s.Rooms))
}

func (s *Server) MatchRoom() *Room {
	for _, room := range s.AvailableRooms {
		if !room.IsFull() {
			return room
		}
	}
	room := NewRoom(roomId)
	roomId++
	s.AddRoom(room)
	s.AvailableRooms[room.ID] = room
	return room
}

func (s *Server) GetRoomByPlayerId(pid int32) *Room {
	room, exists := s.Rooms[PlayerRoomID[pid]]
	if !exists {
		return nil
	}
	return room
}

func (s *Server) handlePacket(packet []byte, clientAddr *net.UDPAddr) {
	if len(packet) == 0 {
		return
	}

	opCode := packet[0]

	switch opCode {
	case OpJoin:

		room := s.MatchRoom()

		playerID := playerId
		playerId++

		player := NewPlayer(clientAddr, playerID)
		room.AddPlayer(player)
		PlayerRoomID[playerID] = room.ID

		if room.IsFull() {
			fmt.Println("Server: Room is full, removing from available rooms")
			delete(s.AvailableRooms, room.ID)
		}

		fmt.Printf("Server: Player %d joined room %d\n", playerID, room.ID)
		s.Sender.SendTo(
			player.Addr,
			EncodeAcceptPacket(
				&AcceptPacket{PlayerID: player.ID},
			))

	case OpLocationUpdate:
		state, err := DecodeLocationPacket(packet)
		if err != nil {
			fmt.Printf("Server: Error decoding location packet from %s: %v\n", clientAddr, err)
			return
		}

		if len(s.Rooms) > 0 {
			room := s.GetRoomByPlayerId(state.PlayerID)

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

	case OpHit:
		hitPacket, err := DecodeHitPacket(packet)
		if err != nil {
			fmt.Printf("Server: Error decoding hit packet from %s: %v\n", clientAddr, err)
			return
		}

		if len(s.Rooms) > 0 {
			room := s.GetRoomByPlayerId(hitPacket.PlayerID)

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
			room := s.GetRoomByPlayerId(shootPacket.PlayerID)

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
	}
}
