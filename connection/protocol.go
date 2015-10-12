package connection

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

type Packet struct {
	Type     string //REQ|RSP
	Identity uint32
	BodySize uint32
	Body     []byte //用户数据
}

var (
	ErrProtoUnknownType     = errors.New("protocol: unknown type")
	ErrProtoBadPacket       = errors.New("bad packet")
	ErrProtoBadPacketLength = errors.New("bad packet: not enough packet length")
	ErrProtoBadBodyLength   = errors.New("bad packet: not enough body length")
)

/*
   [REQ|RSP][32-bit identity][32-bit body-size][Y-bit body]
   [   3   ][       4       ][      4         ][    Y     ]
*/
func (p *Packet) encode() ([]byte, error) {
	if p.Type != "REQ" && p.Type != "RSP" {
		return nil, ErrProtoUnknownType
	}

	body_size := uint32(len(p.Body))
	data := make([]byte, 0, 11+body_size)

	data = append(data, []byte(p.Type)...)

	identity_sl := make([]byte, 4)
	binary.BigEndian.PutUint32(identity_sl, p.Identity)
	data = append(data, identity_sl...)

	body_size_sl := make([]byte, 4)
	binary.BigEndian.PutUint32(body_size_sl, body_size)
	data = append(data, body_size_sl...)

	data = append(data, p.Body...)
	data = append(data, []byte{'\r', '\r', '\n'}...)

	return data, nil
}

func (p *Packet) String() string {
	return fmt.Sprintf("<type: %s, id: %d, body_size: %d>", p.Type, p.Identity, p.BodySize)
}

/*
   [REQ|RSP][32-bit identity][32-bit body-size][Y-bit body]
   [   3   ][       4       ][      4         ][    Y     ]
*/
func decode_packet(data []byte) (p *Packet, err error) {
	if len(data) < 11 {
		return nil, ErrProtoBadPacketLength
	}

	p = &Packet{}
	p.Type = string(data[0:3])
	if p.Type != "REQ" && p.Type != "RSP" {
		err = ErrProtoUnknownType
		return nil, err
	}

	_ = binary.Read(bytes.NewReader(data[3:7]), binary.BigEndian, &p.Identity)
	_ = binary.Read(bytes.NewReader(data[7:11]), binary.BigEndian, &p.BodySize)

	p.Body = data[11:]

	if uint32(len(p.Body)) != p.BodySize {
		log.Printf("id=%d, bodysize=%d, body=|%s|", p.Identity, p.BodySize, string(p.Body))
		err = ErrProtoBadBodyLength
		return nil, err
	}

	return p, nil
}
