package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"log/slog"
	"net"
	"os"

	"github.com/multiformats/go-varint"
)

type ConnectionState uint8

const (
	Handshake ConnectionState = iota
	Status
	Login
	Transfer
)

type ServerState struct {
	CurrentState ConnectionState
}

func NewServerState() ServerState {
	return ServerState{
		CurrentState: Handshake,
	}
}

func (s *ServerState) changeConnectionState(nextState ConnectionState) {
	s.CurrentState = nextState
	slog.Info("changing state", "newState", nextState)

}

var state = NewServerState()

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	srv, err := net.Listen("tcp", ":25565")
	if err != nil {
		slog.Error("dial error:", "msg", err)
		os.Exit(1)
	}
	slog.Info("started a tcp server on port 25565")

	for {
		conn, err := srv.Accept()
		if err != nil {
			slog.Error("accepting connection", "msg", err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			slog.Error("closing connection", "err", err)
		}

	}(conn)

	reader := bufio.NewReader(conn)

	packetLength, _ := varint.ReadUvarint(reader)
	slog.Debug("packetLength", "val", packetLength)

	buf := make([]byte, packetLength)

	if _, err := reader.Read(buf); err != nil {
		slog.Error("error reading bytes from connection", "msg", err)
		return
	}

	switch state.CurrentState {
	case Handshake:
		slog.Info("handling handshake state case")

		slog.Debug("handle the state")
		bufCp := make([]byte, len(buf))
		copy(bufCp, buf)
		handshakePacket := parseHandshakePacket(bufCp)

		slog.Debug("handshake packet", "value", handshakePacket)
		state.changeConnectionState(handshakePacket.NextState)
	}
}

type HandshakePacket struct {
	Address         string
	ProtocolVersion uint64
	Port            []byte
	NextState       ConnectionState
}

func parseHandshakePacket(data []byte) HandshakePacket {
	var handshakePacket HandshakePacket

	// packet ID takes one byte
	packetID := data[0]
	slog.Debug("handshake", "packet ID", packetID)
	data = cutSlice(data, 1)

	protocolVersion, n := binary.Uvarint(data)
	handshakePacket.ProtocolVersion = protocolVersion
	data = cutSlice(data, n)

	serverAddress, n := String(data)
	handshakePacket.Address = serverAddress
	data = cutSlice(data, n)

	serverPort := data[:2]
	handshakePacket.Port = serverPort
	data = cutSlice(data, 2)

	nextState := ConnectionState(data[0])
	handshakePacket.NextState = nextState

	return handshakePacket
}

var ErrStringTooBig = errors.New("String is too big")

func String(buf []byte) (val string, n int) {
	stringLength, n := binary.Uvarint(buf)
	buf = cutSlice(buf, n)

	stringPart := buf[:stringLength]

	return string(stringPart), len(stringPart) + n

}

func cutSlice(slice []byte, offset int) []byte {
	return slice[offset:]
}
