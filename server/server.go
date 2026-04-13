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
	delete(s.Rooms, roomID)
	fmt.Printf("Server: Removed room %d, total rooms: %d\n", roomID, len(s.Rooms))
}

func (s *Server) MatchRoom() *Room {
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
	case OpLocationUpdate:
		s.handleLocationUpdate(packet, clientAddr)
	case OpHit:
		s.handleHit(packet, clientAddr)
	case OpShoot:
		s.handleShoot(packet, clientAddr)
	default:
		fmt.Printf("Server: Received unknown packet from %s: %x\n", clientAddr, packet)
	}
}

func (s *Server) handleJoin(packet []byte, clientAddr *net.UDPAddr) {
	if len(packet) < 5 {
		return
	}

	sessionLen := int(binary.BigEndian.Uint32(packet[1:5]))
	if len(packet) < 5+sessionLen {
		return
	}

	sessionID := string(packet[5 : 5+sessionLen])

	s.mu.Lock()
	session, exists := s.Sessions.Get(sessionID)
	if !exists {
		fmt.Printf("Server: Invalid session ID from %s: %s\n", clientAddr, sessionID)
		s.mu.Unlock()
		return
	}

	session.LastActive = time.Now()
	fmt.Printf("Server: Valid Player %s Joined, Sending PlayerID %d\n", sessionID, s.PlayerIdCounter)

	room := s.MatchRoom()
	playerID := s.PlayerIdCounter
	s.PlayerIdCounter++

	player := NewPlayer(clientAddr, playerID)
	room.AddPlayer(player)
	s.PlayerRoomID[playerID] = room.ID

	if room.IsFull() {
		fmt.Println("Server: Room is full, removing from available rooms")
		delete(s.AvailableRooms, room.ID)
	}

	s.mu.Unlock()

	fmt.Printf("Server: Player %d joined room %d\n", playerID, room.ID)
	s.Sender.SendTo(
		player.Addr,
		EncodeAcceptPacket(
			&AcceptPacket{PlayerID: player.ID},
		))
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
	room.Engine.UpdateHp(target, -hitPacket.Damage)

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
