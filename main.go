package main

import (
	"bytes"
	"io"
	"log/slog"
	"net"
	"os"
)

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
