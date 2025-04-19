package handshake

import (
	"bufio"
	"bytes"
	"errors"
	"log/slog"

	"github.com/multiformats/go-varint"
	"github.com/perzhul/Minerva/protocol"
)

type HandshakePacket struct {
	Address         string
	ProtocolVersion uint64
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

	protocolVersion, err := varint.ReadUvarint(r)
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
