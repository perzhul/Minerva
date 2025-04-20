package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"os"

	"github.com/multiformats/go-varint"
	"github.com/perzhul/Minerva/config"
	"github.com/perzhul/Minerva/protocol"
	"github.com/perzhul/Minerva/protocol/handshake"
	"github.com/perzhul/Minerva/protocol/login"
	"github.com/perzhul/Minerva/protocol/ping"
	"github.com/perzhul/Minerva/protocol/status"
)

const (
	Handshaking protocol.ConnectionState = iota
	Status
	Login
	Transfer
)

type ServerState struct {
	cfg           *config.ServerConfig
	OnlinePlayers uint64
}

func main() {
	slog := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))

	srv, err := net.Listen("tcp", ":25565")
	if err != nil {
		slog.Error("dial error:", "msg", err)
		os.Exit(1)
	}
	slog.Info("started a tcp server on port 25565")

	config := &config.ServerConfig{
		Version:    "1.25.1",
		MaxPlayers: 50,
	}

	state := &ServerState{
		OnlinePlayers: 1,
		cfg:           config,
	}

	for {
		conn, err := srv.Accept()
		if err != nil {
			slog.Error("accepting connection", "msg", err)
		}

		go handleConnection(conn, state)
	}
}

func handleConnection(conn net.Conn, state *ServerState) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			slog.Error("closing connection", "err", err)
		}

	}(conn)

	ctx := &ClientContext{
		conn:   conn,
		reader: bufio.NewReader(conn),
		state:  Handshaking,
	}

	for {
		if err := ctx.handleNextPacket(state); err != nil {
			slog.Error("connection error", "msg", err)
			return
		}
	}

}

type ClientContext struct {
	conn   net.Conn
	reader *bufio.Reader
	state  protocol.ConnectionState
}

func (ctx *ClientContext) changeState(newState protocol.ConnectionState) {
	oldState := ctx.state
	ctx.state = newState
	slog.Info(
		"state change",
		"old state", oldState,
		"new state", ctx.state,
	)
}

func (ctx *ClientContext) handleNextPacket(state *ServerState) error {
	packetLength, err := handshake.ReadVarInt(ctx.reader)
	if err != nil {
		return err
	}

	buf := make([]byte, packetLength)

	if _, err := ctx.reader.Read(buf); err != nil {
		return errors.Join(errors.New("error reading from the connection"), err)
	}

	packetID := buf[0] // because the length prefix is already read from the buffer

	switch ctx.state {
	case Handshaking:
		slog.Debug("handling handshake state case")

		handshakePacket, err := handshake.ParseHandshakePacket(buf)
		state.cfg.ProtocolVersion = handshakePacket.ProtocolVersion

		if err != nil {
			return errors.New("error parsing handshake packet")
		}

		slog.Debug("handshake packet", "value", handshakePacket)

		ctx.changeState(handshakePacket.NextState)

		slog.Debug("changed state", "current state", ctx.state)

	case Status:
		slog.Debug("handling the status case")

		switch byte(packetID) {
		case status.StatusResponsePacketID:
			slog.Debug("handling status response packet", "packet ID", packetID)
			if err := ctx.handleStatusResponsePacket(state); err != nil {
				slog.Error("handling status response packet")
				return err
			}

		case ping.PingPacketID:
			slog.Debug("handling the ping request packet", "packet ID", packetID)
			if err := ctx.handlePingPacket(buf); err != nil {
				slog.Error("handling ping packet")
				return err
			}
		}
	case Login:
		slog.Debug("handling login packet...")
		if err := ctx.handleLoginPacket(buf); err != nil {
			slog.Error("handling login packet")
			return err
		}
	}

	return nil
}

func (ctx *ClientContext) handleLoginPacket(data []byte) error {
	loginPacket, err := login.ParseLoginPacket(data)
	if err != nil {
		return err
	}

	slog.Info("parsed login packet",
		"username", loginPacket.Name,
		"user UUID", loginPacket.UUID,
	)

	return nil
}

func (ctx *ClientContext) handlePingPacket(data []byte) error {
	packetBuf := []byte{}
	packetBuf = append(packetBuf, varint.ToUvarint(uint64(len(data)))...)
	packetBuf = append(packetBuf, data...)

	if _, err := ctx.conn.Write(packetBuf); err != nil {
		return err
	}

	return nil
}

func (ctx *ClientContext) handleStatusResponsePacket(state *ServerState) error {
	favicon, err := encodeImageToFavicon(config.FAVICON_PATH)
	if err != nil {
		return err
	}

	statusResponse := status.Response{
		Version: status.Version{
			Name:     state.cfg.Version,
			Protocol: state.cfg.ProtocolVersion,
		},
		Players: &status.Players{
			Max:    state.cfg.MaxPlayers,
			Online: state.OnlinePlayers,
		},
		Description: &status.Description{
			Text: "Hello, world!",
		},
		Favicon:            favicon,
		EnforcesSecureChat: false,
	}

	jsonData, err := json.Marshal(statusResponse)
	if err != nil {
		return err
	}

	packetID := status.StatusResponsePacketID

	jsonLengthPrefix := varint.ToUvarint(uint64(len(jsonData)))

	payload := []byte{}

	payload = append(payload, packetID)
	payload = append(payload, jsonLengthPrefix...)
	payload = append(payload, jsonData...)

	payloadLengthPrefix := varint.ToUvarint(uint64(len(payload)))

	data := append(payloadLengthPrefix, payload...)

	if _, err := ctx.conn.Write(data); err != nil {
		return err
	}

	return nil

}
