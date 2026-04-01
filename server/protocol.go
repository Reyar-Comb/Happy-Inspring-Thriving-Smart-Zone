package server

import (
	"encoding/binary"
	"errors"
)

const (
	OpJoin           byte = 0x01
	OpAccept         byte = 0x02
	OpLocationUpdate byte = 0x03
	OpHpUpdate       byte = 0x04
	OpShoot          byte = 0x05
	OpHit            byte = 0x06
)

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

type AcceptPacket struct {
	PlayerID int32
}

var ErrInvalidPacket = errors.New("invalid packet")

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

func EncodeAcceptPacket(packet *AcceptPacket) []byte {
	buf := make([]byte, 5)
	buf[0] = OpAccept
	binary.BigEndian.PutUint32(buf[1:5], uint32(packet.PlayerID))
	return buf
}
