package main

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"time"

	"gopkg/connection"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:5555")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			panic(err)
		}

		fmt.Println("Connected from", conn.RemoteAddr().String())

		sock := connection.NewSocket(conn)
		c := connection.NewConnection(sock, 10240, &time_handler{}, nil)

		go func(conn connection.Connection) {
			defer func() {
				fmt.Println("Disconnected from")
				conn.Close()
			}()

			cc := make(chan bool)
			<-cc
		}(c)

	}
}

type time_handler struct{}

func (th *time_handler) ProcessRequest(command []byte) ([]byte, error) {
	switch string(command) {
	case "time":
		return []byte(time.Now().String()), nil
	default:
		return nil, errors.New("unkown command")
	}
}

func (th *time_handler) ProcessOrphanResponse(data []byte) error {
	return connection.ErrOrphanRespDiscard
}
