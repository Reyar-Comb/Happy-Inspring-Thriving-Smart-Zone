package server

import (
	"encoding/binary"
	"errors"
)

const (
	OpJoin           byte = 0x01 //Client -> Server
	OpJoinAck        byte = 0x02 //Server -> Client
	OpLocationUpdate byte = 0x03 //Client -> Server, Server -> Client
	OpHpUpdate       byte = 0x04 //Server -> Client
	OpShoot          byte = 0x05 //Client -> Server, Server -> Client
	OpHit            byte = 0x06 //Client -> Server
	OpOver           byte = 0x07 //Server -> Client
	OpReady          byte = 0x08 //Client -> Server
	OpReadyAck       byte = 0x09 //Server -> Client
	OpLeave          byte = 0x0A //Client -> Server
	OpLeaveAck       byte = 0x0B //Server -> Client
)

const (
	StWaiting byte = 0x00
	StReady   byte = 0x01
)

type JoinPacket struct {
	SessionID string
	RoomCode  int32
}

type JoinAckPacket struct {
	PlayerID int32
	RoomCode int32
	State    byte
}

type LocationPacket struct {
	PlayerID int32
	X        int32
	Y        int32
}

type HpPacket struct {
	PlayerID int32
	Hp       int32
}

type ShootPacket struct {
	PlayerID int32
	X        int32
	Y        int32
	SpeedX   int32
	SpeedY   int32
	Power    int32
}

type HitPacket struct {
	PlayerID int32
	Damage   int32
}

type OverPacket struct {
	WinnerPlayerID int32
}

type ReadyPacket struct {
	PlayerID int32
}

type ReadyAckPacket struct {
	PlayerID int32
	State    byte
}

type LeavePacket struct {
	PlayerID int32
}

type LeaveAckPacket struct {
	PlayerID int32
}

type CreateRoomAckPacket struct {
	RoomCode int32
	PlayerID int32
}

var ErrInvalidPacket = errors.New("invalid packet")

func EncodeJoinAckPacket(packet *JoinAckPacket) []byte {
	buf := make([]byte, 13)
	buf[0] = OpJoinAck
	binary.BigEndian.PutUint32(buf[1:5], uint32(packet.PlayerID))
	binary.BigEndian.PutUint32(buf[5:9], uint32(packet.RoomCode))
	buf[9] = packet.State
	return buf
}

func DecodeReadyPacket(data []byte) (*ReadyPacket, error) {
	if len(data) < 5 {
		return nil, ErrInvalidPacket
	}
	return &ReadyPacket{
		PlayerID: int32(binary.BigEndian.Uint32(data[1:5])),
	}, nil
}

func EncodeReadyAckPacket(packet *ReadyAckPacket) []byte {
	buf := make([]byte, 5)
	buf[0] = OpReadyAck
	binary.BigEndian.PutUint32(buf[1:5], uint32(packet.PlayerID))
	buf[5] = packet.State
	return buf
}

func DecodeLocationPacket(data []byte) (*LocationPacket, error) {
	if len(data) < 13 {
		return nil, ErrInvalidPacket
	}
	return &LocationPacket{
		PlayerID: int32(binary.BigEndian.Uint32(data[1:5])),
		X:        int32(binary.BigEndian.Uint32(data[5:9])),
		Y:        int32(binary.BigEndian.Uint32(data[9:13])),
	}, nil

}

func EncodeLocationPacket(packet *LocationPacket) []byte {
	buf := make([]byte, 13)
	buf[0] = OpLocationUpdate
	binary.BigEndian.PutUint32(buf[1:5], uint32(packet.PlayerID))
	binary.BigEndian.PutUint32(buf[5:9], uint32(packet.X))
	binary.BigEndian.PutUint32(buf[9:13], uint32(packet.Y))
	return buf
}

func EncodeHpPacket(packet *HpPacket) []byte {
	buf := make([]byte, 9)
	buf[0] = OpHpUpdate
	binary.BigEndian.PutUint32(buf[1:5], uint32(packet.PlayerID))
	binary.BigEndian.PutUint32(buf[5:9], uint32(packet.Hp))
	return buf
}

func DecodeHitPacket(data []byte) (*HitPacket, error) {
	if len(data) < 9 {
		return nil, ErrInvalidPacket
	}
	return &HitPacket{
		PlayerID: int32(binary.BigEndian.Uint32(data[1:5])),
		Damage:   int32(binary.BigEndian.Uint32(data[5:9])),
	}, nil
}

func DecodeShootPacket(data []byte) (*ShootPacket, error) {
	if len(data) < 25 {
		return nil, ErrInvalidPacket
	}
	return &ShootPacket{
		PlayerID: int32(binary.BigEndian.Uint32(data[1:5])),
		X:        int32(binary.BigEndian.Uint32(data[5:9])),
		Y:        int32(binary.BigEndian.Uint32(data[9:13])),
		SpeedX:   int32(binary.BigEndian.Uint32(data[13:17])),
		SpeedY:   int32(binary.BigEndian.Uint32(data[17:21])),
		Power:    int32(binary.BigEndian.Uint32(data[21:25])),
	}, nil
}

func EncodeShootPacket(packet *ShootPacket) []byte {
	buf := make([]byte, 25)
	buf[0] = OpShoot
	binary.BigEndian.PutUint32(buf[1:5], uint32(packet.PlayerID))
	binary.BigEndian.PutUint32(buf[5:9], uint32(packet.X))
	binary.BigEndian.PutUint32(buf[9:13], uint32(packet.Y))
	binary.BigEndian.PutUint32(buf[13:17], uint32(packet.SpeedX))
	binary.BigEndian.PutUint32(buf[17:21], uint32(packet.SpeedY))
	binary.BigEndian.PutUint32(buf[21:25], uint32(packet.Power))
	return buf
}

func EncodeOverPacket(packet *OverPacket) []byte {
	buf := make([]byte, 5)
	buf[0] = OpOver
	binary.BigEndian.PutUint32(buf[1:5], uint32(packet.WinnerPlayerID))
	return buf
}

func DecodeLeavePacket(data []byte) (*LeavePacket, error) {
	if len(data) < 5 {
		return nil, ErrInvalidPacket
	}
	return &LeavePacket{
		PlayerID: int32(binary.BigEndian.Uint32(data[1:5])),
	}, nil
}

func EncodeLeaveAckPacket(packet *LeaveAckPacket) []byte {
	buf := make([]byte, 5)
	buf[0] = OpLeaveAck
	binary.BigEndian.PutUint32(buf[1:5], uint32(packet.PlayerID))
	return buf
}
