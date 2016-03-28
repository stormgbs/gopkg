package connection

/*
   所有的请求必须有应答。
   超时后丢弃。
*/

import (
	"bytes"
	"errors"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var DebugID uint32 = 10

var ErrExited = errors.New("exited")
var ErrTimeout = errors.New("query time out")
var ErrOutChanWriteTimeout = errors.New("write out-channel time out")
var ErrAppNotFound = errors.New("applicant not found")

var c net.Conn

type Connection interface {
	//local addrs
	LocalAddr() net.Addr

	//remote addr
	RemoteAddr() net.Addr

	//Query send request and waiting for response until timeout.
	Query(req []byte, timeout_ms int64) (resp []byte, err error)

	//Send just send request and return immediately.
	Send(req []byte) error

	//If resp found no app, we forward it to outer channel.
	// SetOutChannel(ch chan<- []byte)
	// SetOutChannelWriteTimeout(d time.Duration)

	//Close and destory conn.
	Close()
}

type connection struct {
	sync.RWMutex

	conn       Socket
	wg         sync.WaitGroup
	applicants map[uint32]*recv_chan
	chrecv     chan *recv_chan
	chsend     chan []byte

	dh DataHandler
	eh ErrorHandler

	identity uint32

	chexit chan bool
	closed bool
}

type recv_chan struct {
	ch    chan []byte
	timer *time.Timer
}

func NewConnection(sock Socket, count int, dh DataHandler, eh ErrorHandler) Connection {
	maxcount := count
	if maxcount < 1024 {
		maxcount = 1024
	}

	chrecv := make(chan *recv_chan, maxcount)
	for i := 0; i < maxcount; i++ {
		chrecv <- &recv_chan{ch: make(chan []byte, 1), timer: &time.Timer{}}
	}

	c := &connection{
		conn:       sock,
		applicants: make(map[uint32]*recv_chan),
		chrecv:     chrecv,
		chsend:     make(chan []byte, maxcount),
		// out_channel: out_channel,
		dh:     dh,
		eh:     eh,
		chexit: make(chan bool),
	}

	if dh == nil {
		c.dh = &default_data_handler{}
	}

	if eh == nil {
		c.eh = &default_error_handler{}
	}

	go c.start()

	return c
}

func (c *connection) start() {
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
			c.eh.OnError(err)
		}
	}
	close(errch)
}

// func (c *connection) SetOutChannel(ch chan<- []byte) {
// 	c.out_channel = ch
// }

// func (c *connection) SetOutChannelWriteTimeout(d time.Duration) {
// 	c.out_channel_w_tmout = d
// }

func (c *connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}
func (c *connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *connection) Send(data []byte) error {
	id := c.newIdentity()
	return c.write_request(id, data)
}

func (c *connection) Close() {
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
func (c *connection) Query(data []byte, timeout_ms int64) (res []byte, err error) {
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

	if timeout_ms <= 0 {
		timeout_ms = 5000 //5s
	}
	recv.timer = time.NewTimer(time.Duration(timeout_ms) * time.Millisecond)

	// log.Println("write_request ...")
	err = c.write_request(id, data)
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

func (c *connection) write_request(identity uint32, data []byte) (err error) {
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

func (c *connection) write(data []byte) error {
	select {
	case <-c.chexit:
		return ErrExited
	case c.chsend <- data:
		break
	}
	return nil
}

func (c *connection) send() (err error) {
	defer func() {
		if err != nil {
			log.Println("Connection send ", c.RemoteAddr().String(), ", error:", err)
		}
	}()

	for {
		select {
		case <-c.chexit:
			err = ErrExited
			return
		case data := <-c.chsend:
			if err = c.conn.Write(data); err != nil {
				return
			}
		}
	}
}

func (c *connection) recv() (err error) {
	defer func() {
		if err != nil {
			log.Println("Connection recv ", c.RemoteAddr().String(), ", error:", err)
		}
	}()

	var pre []byte = make([]byte, 0, 16384)

	for {
		select {
		case <-c.chexit:
			err = ErrExited
			return
		default:
		}

		var src []byte
		src, err = c.conn.Read()
		if err != nil {
			return
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

func (c *connection) handle(data []byte) (err error) {
	var pkt *Packet
	pkt, err = decode_packet(data)
	if err != nil {
		log.Println("decode_packet error:", err)
		return
	}

	switch pkt.Type {
	case "REQ":
		return c.process_request_packet(pkt)
		// break
	case "RSP":
		return c.process_response_packet(pkt)
		// break
	default:
		err = ErrProtoUnknownType
	}

	return
}

// 处理 对方的请求
func (c *connection) process_request_packet(p *Packet) (err error) {
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
func (c *connection) process_response_packet(p *Packet) (err error) {
	if p == nil {
		return errors.New("empty packet")
	}

	recv, ok := c.popApplicant(p.Identity)
	if !ok {
		return c.dh.ProcessOrphanResponse(p.Body)
	} else if recv == nil {
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

// func (c *connection) push_out_channel(d []byte) error {
// 	if c.out_channel_w_tmout <= 0 {
// 		for {
// 			select {
// 			case c.out_channel <- d:
// 				return nil
// 			}
// 		}
// 	} else {
// 		timer := time.NewTimer(c.out_channel_w_tmout)
// 		for {
// 			select {
// 			case c.out_channel <- d:
// 				return nil
// 			case <-timer.C:
// 				return ErrOutChanWriteTimeout
// 			}
// 		}
// 	}
// }

func (c *connection) addApplicant(id uint32, recv *recv_chan) {
	c.Lock()
	c.applicants[id] = recv
	c.Unlock()
}

func (c *connection) popApplicant(id uint32) (*recv_chan, bool) {
	c.Lock()
	defer c.Unlock()

	rv, ok := c.applicants[id]
	if !ok {
		return nil, false
	}
	delete(c.applicants, id)

	return rv, true
}

func (c *connection) newIdentity() uint32 {
	return atomic.AddUint32(&c.identity, 1)
}
