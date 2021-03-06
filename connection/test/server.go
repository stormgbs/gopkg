package main

import (
	"fmt"
	"net"
	"time"
    "errors"
    "runtime"

    "github.com/stormgbs/gopkg/connection"
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

        sock := connection.NewTcpSocket(conn)
        c := connection.NewConnection(sock, 10240, &time_handler{}, nil, 5*time.Second)

		go func(conn *connection.Connection){
            defer func() {
                fmt.Println("Disconnected from")
                conn.Close()
            }()

            cc := make(chan bool)
            <- cc

			//for {
            //    _, err := conn.Write([]byte(time.Now().String()))
            //    //rsp, err := conn.Query([]byte(time.Now().String()))
            //    if err != nil {
            //        return
            //    }

            //    time.Sleep(time.Second)
			//}

		}(c)

	}
}

type time_handler struct {}

func (th *time_handler) ProcessRequest(command []byte) ([]byte, error) {
    switch string(command) {
    case "time":
        return []byte(time.Now().String()), nil
    default:
        return nil, errors.New("unkown command")
    }
}
