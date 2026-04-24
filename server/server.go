package server

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Reyar-Comb/HITPlane/config"
)

type Server struct {
	mu sync.RWMutex

	// UDP
	Addr   string
	Conn   *net.UDPConn
	Sender *Sender

	// Game state
	Rooms           map[int32]*Room
	AvailableRooms  map[int32]*Room
	PlayerRoomID    map[int32]int32
	PlayerIdCounter int32
	RoomIdCounter   int32

	// HTTP
	Users    *UserStore
	Sessions *SessionManager
}

func NewServer() *Server {
	return &Server{
		Rooms:           make(map[int32]*Room),
		AvailableRooms:  make(map[int32]*Room),
		PlayerRoomID:    make(map[int32]int32),
		PlayerIdCounter: 1001,
		RoomIdCounter:   1,
		Users:           NewUserStore(),
		Sessions:        NewSessionManager(),
	}
}

func (s *Server) StartUDP() error {
	udpAddr, err := net.ResolveUDPAddr("udp", config.GlobalConfig.UDPPort)
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

	fmt.Println("Server: UDP server listening on", config.GlobalConfig.UDPPort)

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

// ----------------------------------------------------------
// Room management
// ----------------------------------------------------------
func (s *Server) AddRoom(room *Room) {
	s.Rooms[room.ID] = room
	fmt.Printf("Server: Added new room, total rooms: ")
	for id := range s.Rooms {
		fmt.Printf("%d ", id)
	}
	fmt.Printf("\n")
}

func (s *Server) RemoveRoom(roomID int32) {
	delete(s.AvailableRooms, roomID)
	delete(s.Rooms, roomID)
	fmt.Printf("Server: Removed room %d, total rooms: %d\n", roomID, len(s.Rooms))
}

func (s *Server) MatchRoom(roomCode int32) *Room {
	switch roomCode {

	case 0: // 0 -> auto match
		for _, room := range s.AvailableRooms {
			if !room.IsFull() {
				return room
			}
		}
		room := NewRoom(s.RoomIdCounter)
		s.RoomIdCounter++
		s.AddRoom(room)
		s.AvailableRooms[room.ID] = room
		return room

	case 1: // 1 -> create room
		room := NewRoom(s.RoomIdCounter)
		s.RoomIdCounter++
		s.AddRoom(room)
		s.AvailableRooms[room.ID] = room
		return room

	default: // specific room code
		for _, room := range s.AvailableRooms {
			if room.RoomCode == roomCode && !room.IsFull() {
				return room
			}
		}
		return nil
	}
}

func (s *Server) GetRoomByPlayerId(pid int32) *Room {
	room, exists := s.Rooms[s.PlayerRoomID[pid]]
	if !exists {
		return nil
	}
	return room
}

// ----------------------------------------------------------
// Packet handling
// ----------------------------------------------------------
func (s *Server) handlePacket(packet []byte, clientAddr *net.UDPAddr) {
	if len(packet) == 0 {
		return
	}

	opCode := packet[0]

	switch opCode {
	case OpJoin:
		s.handleJoin(packet, clientAddr)
	case OpReady:
		s.handleReady(packet, clientAddr)
	case OpLocationUpdate:
		s.handleLocationUpdate(packet, clientAddr)
	case OpHit:
		s.handleHit(packet, clientAddr)
	case OpShoot:
		s.handleShoot(packet, clientAddr)
	case OpLeave:
		s.handleLeave(packet, clientAddr)
	default:
		fmt.Printf("Server: Received unknown packet from %s: %x\n", clientAddr, packet)
	}
}

func (s *Server) handleJoin(packet []byte, clientAddr *net.UDPAddr) {
	if len(packet) < 9 {
		return
	}

	sessionLen := int(binary.BigEndian.Uint32(packet[1:5]))
	if len(packet) < 9+sessionLen {
		return
	}

	roomCode := int32(binary.BigEndian.Uint32(packet[5:9]))

	sessionID := string(packet[9 : 9+sessionLen])

	s.mu.Lock()
	session, exists := s.Sessions.Get(sessionID)
	if !exists {
		fmt.Printf("Server: Invalid session ID from %s: %s\n", clientAddr, sessionID)
		s.mu.Unlock()
		return
	}

	session.LastActive = time.Now()
	fmt.Printf("Server: Valid Player %s Joined, Sending PlayerID %d\n", sessionID, s.PlayerIdCounter)

	room := s.MatchRoom(roomCode)
	if room == nil {
		fmt.Printf("Server: No available room for player %s with room code %d\n", sessionID, roomCode)
		s.mu.Unlock()

		s.Sender.SendTo(clientAddr, EncodeJoinAckPacket(
			&JoinAckPacket{
				PlayerID: -1,
				RoomCode: -1,
			},
		))
		return
	}
	playerID := s.PlayerIdCounter
	s.PlayerIdCounter++

	player := NewPlayer(clientAddr, playerID)
	room.AddPlayer(player)
	s.PlayerRoomID[playerID] = room.ID
	player1 := playerID
	player1Ready := byte(0x00)
	player2 := int32(-1)
	player2Ready := byte(0x00)
	if len(room.Players) == 2 {
		for id := range room.Players {
			if id != playerID {
				player2 = id
				break
			}
		}
	}

	s.mu.Unlock()

	fmt.Printf("Server: Player %d joined room %d\n", playerID, room.ID)
	s.Sender.SendTo(
		clientAddr,
		EncodeJoinAckPacket(
			&JoinAckPacket{
				PlayerID: player.ID,
				RoomCode: room.RoomCode,
			},
		))
	s.Sender.RoomBroadcast(
		room,
		EncodeRoomUpdatePacket(
			&RoomUpdatePacket{
				PlayerID1: player1,
				Ready1:    player1Ready,
				PlayerID2: player2,
				Ready2:    player2Ready,
			},
		),
	)
}

func (s *Server) handleReady(packet []byte, clientAddr *net.UDPAddr) {
	if len(packet) < 6 {
		return
	}

	readyPlayerID := int32(binary.BigEndian.Uint32(packet[1:5]))
	isReady := bool(packet[5] == 0x01)
	s.mu.Lock()
	room := s.GetRoomByPlayerId(readyPlayerID)
	if room == nil {
		s.mu.Unlock()
		return
	}

	player, exists := room.Players[readyPlayerID]
	if !exists {
		s.mu.Unlock()
		return
	}
	if isReady {
		player.SetReady()
	} else {
		player.SetUnready()
	}
	pid1, rdy1, pid2, rdy2 := room.GetReadyStatus()

	s.mu.Unlock()

	s.Sender.RoomBroadcast(
		room,
		EncodeRoomUpdatePacket(
			&RoomUpdatePacket{
				PlayerID1: pid1,
				Ready1:    rdy1,
				PlayerID2: pid2,
				Ready2:    rdy2,
			},
		),
	)

}

func (s *Server) handleLocationUpdate(packet []byte, clientAddr *net.UDPAddr) {
	s.mu.RLock()
	location, err := DecodeLocationPacket(packet)
	if err != nil {
		fmt.Printf("Server: Error decoding location packet from %s: %v\n", clientAddr, err)
		s.mu.RUnlock()
		return
	}

	room := s.GetRoomByPlayerId(location.PlayerID)
	player, exists := room.Players[location.PlayerID]
	if !exists {
		s.mu.RUnlock()
		return
	}

	room.Engine.UpdateLocation(player, DecodeLocation(location))

	s.mu.RUnlock()

	s.Sender.PlayerBroadcast(
		room,
		location.PlayerID,
		EncodeLocationPacket(
			&LocationPacket{
				PlayerID: location.PlayerID,
				X:        int32(player.Location.X),
				Y:        int32(player.Location.Y),
			},
		),
	)
}

func (s *Server) handleHit(packet []byte, clientAddr *net.UDPAddr) {
	s.mu.Lock()
	hitPacket, err := DecodeHitPacket(packet)
	if err != nil {
		fmt.Printf("Server: Error decoding hit packet from %s: %v\n", clientAddr, err)
		s.mu.Unlock()
		return
	}

	room := s.GetRoomByPlayerId(hitPacket.PlayerID)
	player, exists := room.Players[hitPacket.PlayerID]
	if !exists {
		s.mu.Unlock()
		return
	}
	target := GetAnotherPlayer(room, player)
	if target == nil {
		s.mu.Unlock()
		return
	}
	alive := room.Engine.UpdateHp(target, -hitPacket.Damage)
	if !alive {
		s.Sender.RoomBroadcast(
			player.Room,
			EncodeOverPacket(
				&OverPacket{WinnerPlayerID: player.ID},
			),
		)
	}

	s.mu.Unlock()

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

func (s *Server) handleShoot(packet []byte, clientAddr *net.UDPAddr) {
	s.mu.RLock()
	shootPacket, err := DecodeShootPacket(packet)
	if err != nil {
		fmt.Printf("Server: Error decoding shoot packet from %s: %v\n", clientAddr, err)
		s.mu.RUnlock()
		return
	}

	room := s.GetRoomByPlayerId(shootPacket.PlayerID)
	_, exists := room.Players[shootPacket.PlayerID]
	if !exists {
		s.mu.RUnlock()
		return
	}

	s.mu.RUnlock()

	s.Sender.PlayerBroadcast(
		room,
		shootPacket.PlayerID,
		EncodeShootPacket(shootPacket),
	)
}

func (s *Server) handleLeave(packet []byte, clientAddr *net.UDPAddr) {
	s.mu.Lock()
	leavePacket, err := DecodeLeavePacket(packet)
	if err != nil {
		fmt.Printf("Server: Error decoding leave packet from %s: %v\n", clientAddr, err)
		s.mu.Unlock()
		return
	}

	room := s.GetRoomByPlayerId(leavePacket.PlayerID)
	player, exists := room.Players[leavePacket.PlayerID]
	if !exists {
		s.mu.Unlock()
		return
	}

	room.RemovePlayer(player.ID)
	delete(s.PlayerRoomID, player.ID)

	pid1, rdy1, pid2, rdy2 := room.GetReadyStatus()

	if room.IsEmpty() {
		s.RemoveRoom(room.ID)
	} else {
		room.State = RoomWaiting
		s.AvailableRooms[room.ID] = room
	}

	s.mu.Unlock()
	s.Sender.RoomBroadcast(
		room,
		EncodeRoomUpdatePacket(
			&RoomUpdatePacket{
				PlayerID1: pid1,
				Ready1:    rdy1,
				PlayerID2: pid2,
				Ready2:    rdy2,
			},
		),
	)
}
