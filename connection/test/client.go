package main

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/stormgbs/gopkg/connection"
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

	sock := connection.NewTcpSocket(conn)
	c := connection.NewConnection(sock, 10240, nil, nil, 5*time.Second)

	//defer func() {
	//	fmt.Println("closed from server")
	//	c.Close()
	//}()
	p := 3
	synch := make(chan bool, p)
	for i := 0; i < p; i++ {
		synch <- true
	}

	count := 1
	for {
		fmt.Println("\ntimes:", count)
		select {
		case <-synch:
			go func(ch chan bool) {

				rsp, err := c.Query([]byte("time"))
				fmt.Println(string(rsp), err)
				ch <- true
			}(synch)
		}

		//time.Sleep(time.Second)

		count += 1

	}

}
