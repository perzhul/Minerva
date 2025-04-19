package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/multiformats/go-varint"
	"github.com/perzhul/Minerva/config"
	"github.com/perzhul/Minerva/protocol"
	"github.com/perzhul/Minerva/protocol/handshake"
	"github.com/perzhul/Minerva/protocol/ping"
	"github.com/perzhul/Minerva/protocol/status"
)

const (
	Handshake protocol.ConnectionState = iota
	Status
	Login
	Transfer
)

type ServerState struct {
	mu            sync.RWMutex
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
		OnlinePlayers: 0,
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
		state:  Handshake,
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
	slog.Debug(
		"state change",
		"old state", oldState,
		"new state", ctx.state,
	)
}

func (ctx *ClientContext) handleNextPacket(state *ServerState) error {
	packetLength, err := varint.ReadUvarint(ctx.reader)
	if err != nil {
		slog.Error("", "msg", err)
		return errors.New("error reading packet length")
	}

	slog.Debug("packetLength", "length", packetLength)

	buf := make([]byte, packetLength)

	if _, err := ctx.reader.Read(buf); err != nil {
		return errors.New("error reading bytes from connection")
	}

	switch ctx.state {
	//TODO: add other cases
	case Handshake:
		slog.Debug("handling handshake state case")

		handshakePacket, err := handshake.ParseHandshakePacket(buf)
		if err != nil {
			return errors.New("error parsing handshake packet")
		}

		slog.Debug("handshake packet", "value", handshakePacket)
		ctx.changeState(handshakePacket.NextState)

	case Status:
		slog.Debug("handling the status case")
		packetID := buf[0]

		switch byte(packetID) {
		case status.StatusResponsePacketID:
			slog.Debug("handling status response packet", "packet ID", packetID)
			ctx.handleStatusResponsePacket(state)
		case ping.PingPacketID:
			slog.Debug("handling the ping request packet", "packet ID", packetID)
			packetBuf := []byte{}
			packetBuf = append(packetBuf, varint.ToUvarint(uint64(len(buf)))...)
			packetBuf = append(packetBuf, buf...)

			_, err = ctx.conn.Write(packetBuf)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (ctx *ClientContext) handleStatusResponsePacket(state *ServerState) {
	favicon, err := encodeImageToFavicon(config.FAVICON_PATH)
	if err != nil {
		slog.Error("error encoding to favicon", "msg", err)
	}

	statusResponse := status.Response{
		Version: status.Version{
			Name:     "1.25.1",
			Protocol: 770,
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
