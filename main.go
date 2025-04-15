package main

import (
	"bytes"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/perzhul/Minerva/protocol/util"
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

func (s *ServerState) ChangeConnectionState(nextState ConnectionState) {
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

	for {
		conn, err := srv.Accept()
		if err != nil {
			slog.Error("accepting connection", "msg", err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	var buf bytes.Buffer

	_, err := io.Copy(&buf, conn)
	if err != nil {
		slog.Error("error reading bytes from connection", "msg", err)
		return
	}

	slog.Info("successfully read data from connection", "data", buf.String())
}
