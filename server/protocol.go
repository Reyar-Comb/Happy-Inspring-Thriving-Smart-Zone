package server

import (
	"encoding/binary"
	"errors"
)

const (
	OpJoin        byte = 0x01
	OpAccept      byte = 0x02
	OpStateUpdate byte = 0x03
)

type StatePacket struct {
	PlayerID int32
	X        int32
	Y        int32
}

type AcceptPacket struct {
	PlayerID int32
}

var ErrInvalidPacket = errors.New("invalid packet")

func DecodeStatePacket(data []byte) (*StatePacket, error) {
	if len(data) < 13 {
		return nil, ErrInvalidPacket
	}
	return &StatePacket{
		PlayerID: int32(binary.BigEndian.Uint32(data[1:5])),
		X:        int32(binary.BigEndian.Uint32(data[5:9])),
		Y:        int32(binary.BigEndian.Uint32(data[9:13])),
	}, nil

}

func EncodeStatePacket(packet *StatePacket) []byte {
	buf := make([]byte, 13)
	buf[0] = OpStateUpdate
	binary.BigEndian.PutUint32(buf[1:5], uint32(packet.PlayerID))
	binary.BigEndian.PutUint32(buf[5:9], uint32(packet.X))
	binary.BigEndian.PutUint32(buf[9:13], uint32(packet.Y))
	return buf
}

func EncodeAcceptPacket(packet *AcceptPacket) []byte {
	buf := make([]byte, 5)
	buf[0] = OpAccept
	binary.BigEndian.PutUint32(buf[1:5], uint32(packet.PlayerID))
	return buf
}
