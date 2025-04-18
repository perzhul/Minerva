package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"os"

	"github.com/multiformats/go-varint"
	"github.com/perzhul/Minerva/protocol"
	"github.com/perzhul/Minerva/protocol/status"
)

type ConnectionState uint8

const (
	Handshake ConnectionState = iota
	Status
	Login
	Transfer
)

func (s ConnectionState) String() string {
	switch s {
	case Handshake:
		return "Handshake"
	case Status:
		return "Status"
	case Login:
		return "Login"
	case Transfer:
		return "Transfer"
	default:
		return "Unknown"
	}
}

type Server struct {
	Version  string // 1.25.1
	Protocol []byte
	// TODO: fill with properties, players, status, etc.
}

var srvData = Server{
	Version:  "1.25.1",
	Protocol: varint.ToUvarint(uint64(770)),
}

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

	ctx := &ClientContext{
		conn:   conn,
		reader: bufio.NewReader(conn),
		state:  Handshake,
	}

	for {
		if err := ctx.handleNextPacket(); err != nil {
			slog.Error("connection error", "msg", err)
			return
		}
	}

}

type ClientContext struct {
	conn   net.Conn
	reader *bufio.Reader
	state  ConnectionState
}

func (ctx *ClientContext) changeState(newState ConnectionState) {
	oldState := ctx.state
	ctx.state = newState
	slog.Debug(
		"state change",
		"old state", oldState,
		"new state", ctx.state,
	)
}

func (ctx *ClientContext) handleNextPacket() error {
	packetLength, err := varint.ReadUvarint(ctx.reader)
	if err != nil {
		slog.Error("", "msg", err)
		return errors.New("error reading packet length")
	}

	slog.Debug("packetLength", "val", packetLength)

	buf := make([]byte, packetLength)

	if _, err := ctx.reader.Read(buf); err != nil {
		return errors.New("error reading bytes from connection")
	}

	switch ctx.state {
	//TODO: add other cases
	case Handshake:
		slog.Debug("handling handshake state case")

		handshakePacket, err := parseHandshakePacket(buf)
		if err != nil {
			return errors.New("error parsing handshake packet")
		}

		slog.Debug("handshake packet", "value", handshakePacket)
		ctx.changeState(handshakePacket.NextState)
	case Status:
		slog.Debug("handling the status case")

		version := status.Version{
			Name:     srvData.Version,
			Protocol: varint.ToUvarint(uint64(770)),
		}

		packetData := struct {
			Version            status.Version `json:"version"`
			EnforcesSecureChat bool           `json:"enforcesSecureChat"`
		}{
			Version:            version,
			EnforcesSecureChat: false,
		}

		jsonData, err := json.Marshal(packetData)
		if err != nil {
			slog.Error("error marshalling struct", "msg", err)
		}

		packetID := status.StatusResponsePacketID

		jsonLengthPrefix := varint.ToUvarint(uint64(len(jsonData)))

		payload := []byte{}

		payload = append(payload, packetID)
		payload = append(payload, jsonLengthPrefix...)
		payload = append(payload, jsonData...)

		payloadLengthPrefix := varint.ToUvarint(uint64(len(payload)))

		data := append(payloadLengthPrefix, payload...)

		n, err := ctx.conn.Write(data)
		if err != nil {
			slog.Error(
				"error writing to connection",
				"msg", err,
				"tried to write", data,
			)
		}
		slog.Info("bytes wrote", "n", n)

	}

	return nil
}

type HandshakePacket struct {
	Address         string
	ProtocolVersion uint64
	Port            []byte
	NextState       ConnectionState
}

func parseHandshakePacket(data []byte) (handshakePacket HandshakePacket, err error) {
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
		NextState:       ConnectionState(nextState),
	}, nil
}

var ErrStringTooBig = errors.New("String is too big")
