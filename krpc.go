package dht

import (
	"fmt"
	"net"

	"github.com/zeebo/bencode"
)

type packet struct {
	typ  packetType
	data packetData
}

// packet type
type packetType int

// all type
const (
	Ping packetType = iota
	FindNode
	GetPeers
	AnnouncePeer
)

// krpc protocol message packet
type packetData struct {
	T string        `bencode:"t"`
	Y string        `bencode:"y"`
	Q string        `bencode:"q"`
	A answer        `bencode:"a"`
	R response      `bencode:"r"`
	E []interface{} `bencode:"e"`
}

type response struct {
	ID     string   `bencode:"id"`
	Token  string   `bencode:"token"`
	Nodes  string   `bencode:"nodes"`
	Values []string `bencode:"values"`
}

type answer struct {
	ID       string `bencode:"id"`
	Port     string `bencode:"port"`
	Token    string `bencode:"token"`
	Target   string `bencode:"target"`
	InfoHash string `bencode:"info_hash"`
}

func (p *packetData) Error() (n int64, s string) {
	if len(p.E) == 2 {
		n = p.E[0].(int64)
		s = p.E[1].(string)
	}
	return
}

func (p *packetData) Type() (typ packetType) {
	switch p.T {
	case "pn":
		typ = Ping
	case "fn":
		typ = FindNode
	case "gp":
		typ = GetPeers
	case "ap":
		typ = AnnouncePeer
	}
	return
}

func (p *packetData) Marshal() ([]byte, error) {
	return bencode.EncodeBytes(p)
}

func (p *packetData) Unmarshal(b []byte) (err error) {
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("panic() when bencode.DecodeBytes")
		}
	}()
	err = bencode.DecodeBytes(b, p)
	return
}

type queryMessage struct {
	T string                 `bencode:"t"`
	Y string                 `bencode:"y"`
	Q string                 `bencode:"q"`
	A map[string]interface{} `bencode:"a"`
}

func newQueryMessage(id, q string, data map[string]interface{}) *queryMessage {
	return &queryMessage{T: id, Y: "q", Q: q, A: data}
}

func sendUDPMessage(conn *net.UDPConn, addrs []*net.UDPAddr, data interface{}) {
	if b, err := bencode.EncodeBytes(data); err == nil {
		for _, addr := range addrs {
			conn.WriteToUDP(b, addr)
		}
	}
}
