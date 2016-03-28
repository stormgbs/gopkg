package connection

import (
	"net"
)

type Socket interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Read() ([]byte, error)
	Write([]byte) error
	Close()
}

const SocketDefaultMaxBufferSize int = 1024 * 1024 //1MB

type socket struct {
	conn        net.Conn
	read_buffer []byte
}

func NewSocket(c net.Conn) Socket {
	if tcpc, ok := c.(*net.TCPConn); ok {
		tcpc.SetKeepAlive(true)
	}
	return &socket{
		conn:        c,
		read_buffer: make([]byte, SocketDefaultMaxBufferSize),
	}
}

func (s *socket) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *socket) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *socket) Read() ([]byte, error) {
	n, err := s.conn.Read(s.read_buffer)
	if err != nil {
		return nil, err
	}

	dst := make([]byte, n, n)
	copy(dst, s.read_buffer[:n])
	return dst, nil
}

func (s *socket) Write(data []byte) error {
	_, err := s.conn.Write(data)
	return err
}

func (s *socket) Close() {
	s.conn.Close()
}
