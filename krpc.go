package dht

import (
	"fmt"

	"github.com/zeebo/bencode"
)

/*
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
*/

/*
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
*/

/*
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
*/

func decodeMessage(b []byte, val interface{}) (err error) {
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("")
		}
	}()
	return bencode.DecodeBytes(b, val)
}

func encodeMessage(msg interface{}) ([]byte, error) {
	return bencode.EncodeBytes(msg)
}

type KadResponse struct {
	Data struct {
		ID     []byte   `bencode:"id"`
		Token  string   `bencode:"token"`
		Nodes  string   `bencode:"nodes"`
		Values []string `bencode:"values"`
	} `bencode:"r"`
}

func (r *KadResponse) Nodes() map[string]string {
	nodes := make(map[string]string)
	for i := 0; i < len(r.Data.Nodes); i += 26 {
		id := r.Data.Nodes[i : i+20]
		addr := r.Data.Nodes[i+20 : i+26]
		nodes[id] = fmt.Sprintf("%d.%d.%d.%d:%d", addr[0], addr[1],
			addr[2], addr[3], (uint16(addr[4])<<8)|uint16(addr[5]))
	}
	return nodes
}

type KadAnswer struct {
	Data struct {
		ID       string `bencode:"id"`
		Port     string `bencode:"port"`
		Token    string `bencode:"token"`
		Target   string `bencode:"target"`
		InfoHash string `bencode:"info_hash"`
	} `bencode:"a"`
}

type KadError struct {
	Data []interface{} `bencode:"e"`
}

func (e *KadError) Value() int64 {
	return e.Data[0].(int64)
}

func (e *KadError) String() string {
	return e.Data[1].(string)
}

type KadMsgHeader struct {
	T string `bencode:"t"`
	Y string `bencode:"y"`
}

type KadMsgType int

const (
	MsgTypeQuery KadMsgType = iota
	MsgTypeReply
	MsgTypeError
)

func (h *KadMsgHeader) TID() string {
	return h.T
}

func (h *KadMsgHeader) Type() (t KadMsgType) {
	switch h.Y {
	case "q":
		t = MsgTypeQuery
	case "r":
		t = MsgTypeReply
	case "e":
		t = MsgTypeError
	}
	return
}

type KadQueryMessage struct {
	T string                 `bencode:"t"`
	Y string                 `bencode:"y"`
	Q string                 `bencode:"q"`
	A map[string]interface{} `bencode:"a"`
}

type KadReplyMessage struct {
	T string                 `bencode:"t"`
	Y string                 `bencode:"y"`
	R map[string]interface{} `bencode:"r"`
}

func NewQueryMessage(id, q string, data map[string]interface{}) []byte {
	msg := KadQueryMessage{id, "q", q, data}
	b, err := bencode.EncodeBytes(&msg)
	if err != nil {
		return nil
	}
	return b
}

func NewReplyMessage(id string, data map[string]interface{}) []byte {
	msg := KadReplyMessage{id, "r", data}
	b, err := bencode.EncodeBytes(&msg)
	if err != nil {
		return nil
	}
	return b
}
