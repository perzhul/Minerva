package handshake

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/perzhul/Minerva/protocol"
	"github.com/perzhul/Minerva/types"
)

type HandshakePacket struct {
	Address         string
	ProtocolVersion types.Varint
	Port            []byte
	NextState       protocol.ConnectionState
}

func ParseHandshakePacket(data []byte) (handshakePacket HandshakePacket, err error) {
	r := bufio.NewReader(bytes.NewReader(data))
	packetID, err := r.ReadByte()
	if err != nil {
		return handshakePacket, err
	}
	slog.Debug("handshake", "packet ID", packetID)

	protocolVersion, err := ReadVarInt(r)
	if err != nil {
		return handshakePacket, err
	}

	serverAddress, err := protocol.String(r)
	if err != nil {
		return handshakePacket, err
	}

	if len(data) < 2 {
		return handshakePacket, errors.New("not enough bytes")
	}

	serverPort := make([]byte, 2)
	if _, err := r.Read(serverPort); err != nil {
		return handshakePacket, err
	}

	nextState, err := r.ReadByte()
	if err != nil {
		return handshakePacket, err
	}

	return HandshakePacket{
		Address:         serverAddress,
		ProtocolVersion: protocolVersion,
		Port:            serverPort,
		NextState:       protocol.ConnectionState(nextState),
	}, nil
}

func ReadVarInt(r io.ByteReader) (types.Varint, error) {
	var num int32
	var shift uint
	for range 5 { // max 5 bytes
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		num |= int32(b&0x7F) << shift
		if (b & 0x80) == 0 {
			return types.Varint(num), nil
		}
		shift += 7
	}
	return 0, fmt.Errorf("VarInt is too big")
}
