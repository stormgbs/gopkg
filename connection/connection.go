package connection

/*
   所有的请求必须有应答。
   超时后丢弃。
*/ 

import (
	"bytes"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

var DebugID uint32 = 10

var ErrExited = errors.New("exited")
var ErrTimeout = errors.New("time out")
var ErrAppNotFound = errors.New("applicant not found")

type Connection struct {
	sync.RWMutex

	conn       Socket
	wg         sync.WaitGroup
	applicants map[uint32]*recv_chan
	chrecv     chan *recv_chan
	chsend     chan []byte

	dh DataHandler
	eh ErrorHandler

	identity uint32

	timeout time.Duration
	chexit  chan bool
	closed  bool
}

type recv_chan struct {
	ch    chan []byte
	timer *time.Timer
}

func NewConnection(conn Socket, count int, dh DataHandler, eh ErrorHandler, timeout time.Duration) *Connection {
	maxcount := count
	if maxcount < 1024 {
		maxcount = 1024
	}

	chrecv := make(chan *recv_chan, maxcount)
	for i := 0; i < maxcount; i++ {
		chrecv <- &recv_chan{ch: make(chan []byte, 1), timer: &time.Timer{}}
	}

	c := &Connection{
		conn:       conn,
		applicants: make(map[uint32]*recv_chan),
		chrecv:     chrecv,
		chsend:     make(chan []byte, maxcount),
		dh:         dh,
		eh:         eh,
		timeout:    timeout,
		chexit:     make(chan bool),
	}

	if dh == nil {
		c.dh = &default_data_handler{}
	}

	if eh == nil {
		c.eh = &default_error_handler{}
	}

	if timeout <= 0 {
		c.timeout = 5 * time.Second
	}

	go c.start()

	return c
}

func (c *Connection) start() {
	c.wg.Add(2)

	errch := make(chan error, 2)

	go func(ch chan error) {
		defer c.wg.Done()
		ch <- c.recv()
	}(errch)

	go func(ch chan error) {
		defer c.wg.Done()
		ch <- c.send()
	}(errch)

	for {
		select {
		case <-c.chexit:
			return
		case err := <-errch:
			c.Lock()
			if !c.closed {
				close(c.chexit)
				c.closed = true
			}
			c.Unlock()
			log.Println("Connection closed, error:", err)
		}
	}
	close(errch)
}

func (c *Connection) Close() {
	c.Lock()
	if !c.closed {
		close(c.chexit)
		c.closed = true
	}
	c.Unlock()
	c.conn.Close()
	c.wg.Wait()
}

// Send request and wait response.
func (c *Connection) Query(data []byte) (res []byte, err error) {
	var recv *recv_chan

	select {
	case <-c.chexit:
		return nil, ErrExited
	case recv = <-c.chrecv:
		defer func() {
			recv.timer.Stop()
			c.chrecv <- recv
		}()
	}

	id := c.newIdentity()

	c.addApplicant(id, recv)
	defer c.popApplicant(id)

	recv.timer = time.NewTimer(c.timeout)

	// log.Println("do_request ...")
	err = c.do_request(id, data)
	if err != nil {
		log.Printf("Connection::Query() error: %s", err)
		return nil, err
	}

	select {
	case <-c.chexit:
		return nil, ErrExited
	case <-recv.timer.C:
		return nil, ErrTimeout
	case res = <-recv.ch:
		break
	}
	return res, nil
}

func (c *Connection) do_request(identity uint32, data []byte) (err error) {
	p := Packet{
		Type:     "REQ",
		Identity: identity,
		BodySize: uint32(len(data)),
		Body:     data,
	}

	var pkgbyts []byte
	pkgbyts, err = p.encode()
	if err != nil {
		return err
	}

	if err = c.write(pkgbyts); err != nil {
		return err
	}

	// if p.Identity == DebugID {
	// 	log.Println("xxxx:", p.Identity, p.BodySize, string(p.Body), string(pkgbyts))
	// }

	return nil
}

func (c *Connection) write(data []byte) error {
	select {
	case <-c.chexit:
		return ErrExited
	case c.chsend <- data:
		break
	}
	return nil
}

func (c *Connection) send() error {
	for {
		select {
		case <-c.chexit:
			return ErrExited
		case data := <-c.chsend:
			if err := c.conn.Write(data); err != nil {
				return err
			}
		}
	}
}

func (c *Connection) recv() (err error) {
	defer func() {
		if err != nil {
			select {
			case <-c.chexit:
				err = nil
			default:
				c.eh.OnError(err)
			}
		}
	}()

	var pre []byte = make([]byte, 0, 16384)

	for {
		select {
		case <-c.chexit:
			return nil
		default:
		}

		src, err := c.conn.Read()
		if err != nil {
			return err
		}

		pre = append(pre, src...)

		for {
			m := bytes.Index(pre, []byte{'\r', '\r', '\n'})
			if m >= 0 {
				go c.handle(pre[:m])
				pre = pre[m+3:]
			} else {
				break
			}
		}

	}
	return nil
}

func (c *Connection) handle(data []byte) (err error) {
	var pkt *Packet
	pkt, err = decode_packet(data)
	if err != nil {
		log.Println("decode_packet error:", err)
		return
	}

	switch pkt.Type {
	case "REQ":
		go c.process_request_packet(pkt)
		break
	case "RSP":
		go c.process_response_packet(pkt)
		break
	default:
		err = ErrProtoUnknownType
	}

	return
}

// 处理 对方的请求
func (c *Connection) process_request_packet(p *Packet) (err error) {
	if p == nil {
		return errors.New("empty packet")
	}

	rsp, err := c.dh.ProcessRequest(p.Body)

	rsp_pkt := *p
	rsp_pkt.Type = "RSP"
	rsp_pkt.Body = rsp
	rsp_pkt.BodySize = uint32(len(rsp))

	var rsp_data []byte
	if rsp_data, err = rsp_pkt.encode(); err != nil {
		return err
	}

	if err = c.write(rsp_data); err != nil {
		return err
	}

	return nil
}

// 处理 对方的响应
// 查找recv_chan
func (c *Connection) process_response_packet(p *Packet) (err error) {
	if p == nil {
		return errors.New("empty packet")
	}

	recv, ok := c.popApplicant(p.Identity)
	if !ok || recv == nil {
		return ErrAppNotFound
	}

	// log.Printf("response package: %s|%d|%d", string(p.Body), p.BodySize, len(recv.ch))
	select {
	case <-c.chexit:
		return ErrExited
	case <-recv.timer.C:
		return ErrTimeout
	case recv.ch <- p.Body:
		break
	}

	return nil
}

func (c *Connection) addApplicant(id uint32, recv *recv_chan) {
	c.Lock()
	c.applicants[id] = recv
	c.Unlock()
}

func (c *Connection) popApplicant(id uint32) (*recv_chan, bool) {
	c.Lock()
	defer c.Unlock()

	rv, ok := c.applicants[id]
	if !ok {
		return nil, false
	}
	delete(c.applicants, id)

	return rv, true
}

func (c *Connection) newIdentity() uint32 {
	return atomic.AddUint32(&c.identity, 1)
}
