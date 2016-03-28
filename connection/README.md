connection
===


Connection package implements TCP connection multiplexing. It is easy to use, powerfull and high-performace.

We can send many requests to oppoiste-end in the same time, but it only supports Query(request and wait for response) now.

## Interface&Function

#### Connection

```go
type Connection interface {
    LocalAddr() net.Addr
    RemoteAddr() net.Addr
    Query(data []byte, timeout_ms int64) (resp []byte, err error)
    Send(req []byte) error
    Close()
}
```

#### Socket

```go
type Socket interface {
   LocalAddr() net.Addr
   RemoteAddr() net.Addr
   Read() ([]byte, error)
   Write([]byte) error
   Close()
}
```

#### DataHandler

```go
type DataHandler interface {
    ProcessRequest([]byte) ([]byte, error)
    ProcessOrphanResponse([]byte) error
}
```

#### ErrorHandler

```go
type ErrorHandler interface {
    OnError(error)
}
```


#### NewConnection

    func NewConnection(conn Socket, count int, dh DataHandler, eh ErrorHandler, timeout time.Duration) *Connection

timeout: request timeout  

#### NewTcpSocket

    func NewTcpSocket(c *net.TCPConn) Socket

#### Close
    func (c *Connection) Close()

## Usage

#### Step1
Create a TcpConn, and then make a Socket.

    conn, err := net.DialTcp("tcp", nil, server_addr)
    
    socket := connection.NewTcpSocket(conn)

#### Step2
Create a Connection.

    var c connection.Connection
    c = connection.NewConnection(socket, 10240, nil, nil)
    
    
#### Step3
Now send a request and wait for response with a timer.
    
    var myrequest []byte
    rsp_bytes, err := c.Query(myrequest, 1000) //timeout: 1s

#### Step4
Close it.

    c.Close()
    

## Full Example

Here's a complete, runnable example of a small connection based server.

#### Server
server.go:

```go
package main

import (
    "fmt"
    "net"
    "time"
    "errors"
    "runtime"

    "gaobushuang/gopkg/connection"
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
        c := connection.NewConnection(sock, 10240, &time_handler{}, nil)

                go func(conn connection.Connection){
            defer func() {
                fmt.Println("Disconnected from client.")
                conn.Close()
            }()

            cc := make(chan bool)
            <- cc
        }(c)

        }
}


/// Part: DataHandler 
type time_handler struct {}

func (th *time_handler) ProcessRequest(command []byte) ([]byte, error) {
    switch string(command) {
    case "time":
        return []byte(time.Now().String()), nil
    default:
        return nil, errors.New("unkown command")
    }
}
```
#### Client
client.go:

```go
package main

import (
 "fmt"
 "net"
 "os"
 "runtime"

 "gaobushuang/gopkg/connection"
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
 c := connection.NewConnection(sock, 10240, nil, nil)

 p := 3000
 synch := make(chan bool, p)
 for i := 0; i < p; i++ {
  synch <- true
 }

 for {
  fmt.Println("\ntimes:", count)
  select {
  case <-synch:
   go func(ch chan bool) {
    rsp, err := c.Query([]byte("time"), 1000) //timeout: 1s
    fmt.Println(string(rsp), err)
    ch <- true
   }(synch)
  }

 }
}
```

### Benchmark

#### server: 
CPU:    Intel(R) Xeon(R) CPU E5-1620 v2 @ 3.70GHz x8  
Memory: 32GB

#### client: 
CPU:    Intel(R) Xeon(R) CPU E5-1620 v2 @ 3.70GHz x8  
Memory: 32GB

#### result
QPS on single connection, 17.8w:

    $ grep CST nohup.out  | cut -d. -f1 | sort | uniq -c | sort -nk1
    ...
    176033 2015-09-10 14:44:23
    176310 2015-09-10 14:45:07
    178182 2015-09-10 14:45:05


