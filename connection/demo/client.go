package main

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"gopkg/connection"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	addr, err := net.ResolveTCPAddr("tcp", os.Args[1])
	if err != nil {
		panic(err)
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		panic(err)
	}

	sock := connection.NewSocket(conn)
	c := connection.NewConnection(sock, 10240, nil, nil)

	p := 3000
	synch := make(chan bool, p)
	for i := 0; i < p; i++ {
		synch <- true
	}

	for {

		select {
		case <-synch:
			go func(ch chan bool) {
				rsp, err := c.Query([]byte("time"), 1000) //timeout: 1s
				fmt.Println("Query():", string(rsp), err)

				err = c.Send([]byte("time"))
				fmt.Println("Send() error:", err)

				ch <- true
			}(synch)
		}
		time.Sleep(time.Second * 10)
	}

}
