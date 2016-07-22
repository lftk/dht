package dht

import (
	"bytes"
	"fmt"
	"net"

	"github.com/zeebo/bencode"
)

func DecodeMessage(b []byte, val interface{}) (err error) {
	return decodeMessage(b, val)
}

func EncodeMessage(msg interface{}) ([]byte, error) {
	return encodeMessage(msg)
}

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

type KadMsgHeader struct {
	T []byte `bencode:"t"`
	Y string `bencode:"y"`
}

type KadMsgType int

const (
	QueryMessage KadMsgType = iota
	ReplyMessage
	ErrorMessage
)

func (h *KadMsgHeader) TID() []byte {
	return h.T
}

func (h *KadMsgHeader) Type() (t KadMsgType) {
	switch h.Y {
	case "q":
		t = QueryMessage
	case "r":
		t = ReplyMessage
	case "e":
		t = ErrorMessage
	}
	return
}

type KadErrorMessage struct {
	Data []interface{} `bencode:"e"`
}

func (e *KadErrorMessage) Value() int64 {
	return e.Data[0].(int64)
}

func (e *KadErrorMessage) String() string {
	return e.Data[1].(string)
}

type KadMethodType int

const (
	PingMethod KadMethodType = iota
	FindNodeMethod
	GetPeersMethod
	AnnouncePeerMethod
)

type KadQueryMethod struct {
	Q string `bencode:"q"`
}

func (m *KadQueryMethod) Type() (t KadMethodType) {
	switch m.Q {
	case "ping":
		t = PingMethod
	case "find_node":
		t = FindNodeMethod
	case "get_peers":
		t = GetPeersMethod
	case "announce_peer":
		t = AnnouncePeerMethod
	}
	return
}

type KadQueryMessage struct {
	T []byte                 `bencode:"t"`
	Y string                 `bencode:"y"`
	Q string                 `bencode:"q"`
	A map[string]interface{} `bencode:"a"`
}

type KadReplyMessage struct {
	T []byte                 `bencode:"t"`
	Y string                 `bencode:"y"`
	R map[string]interface{} `bencode:"r"`
}

func NewQueryMessage(tid []byte, q string, data map[string]interface{}) *KadQueryMessage {
	return &KadQueryMessage{tid, "q", q, data}
}

func NewReplyMessage(tid []byte, data map[string]interface{}) *KadReplyMessage {
	return &KadReplyMessage{tid, "r", data}
}

type KadArguments struct {
	Data struct {
		ID       string `bencode:"id"`
		Port     string `bencode:"port"`
		Token    string `bencode:"token"`
		Target   string `bencode:"target"`
		InfoHash string `bencode:"info_hash"`
	} `bencode:"a"`
}

type KadValues struct {
	Data struct {
		ID     []byte   `bencode:"id"`
		Token  string   `bencode:"token"`
		Nodes  string   `bencode:"nodes"`
		Values []string `bencode:"values"`
	} `bencode:"r"`
}

func (v *KadValues) Nodes() map[string]string {
	nodes := make(map[string]string)
	for i := 0; i < len(v.Data.Nodes); i += 26 {
		id := v.Data.Nodes[i : i+20]
		addr := v.Data.Nodes[i+20 : i+26]
		nodes[id] = fmt.Sprintf("%d.%d.%d.%d:%d", addr[0], addr[1],
			addr[2], addr[3], (uint16(addr[4])<<8)|uint16(addr[5]))
	}
	return nodes
}

type KadRequest struct {
	Method string `bencode:"q"`
	Data   struct {
		ID       []byte `bencode:"id"`
		Port     int64  `bencode:"port"`
		Token    string `bencode:"token"`
		Target   string `bencode:"target"`
		InfoHash string `bencode:"info_hash"`
	} `bencode:"a"`
}

func (q *KadRequest) ID() []byte {
	return q.Data.ID
}

/*
func (q *KadRequest) ID() (id *ID) {
	id, _ = NewID([]byte(q.Data.ID))
	return
}
*/

func (q *KadRequest) Port() int64 {
	return q.Data.Port
}

func (q *KadRequest) Token() []byte {
	return []byte(q.Data.Token)
}

func (q *KadRequest) Target() (id *ID) {
	id, _ = NewID([]byte(q.Data.Target))
	return
}

func (q *KadRequest) InfoHash() (id *ID) {
	id, _ = NewID([]byte(q.Data.InfoHash))
	return
}

type KadResponse struct {
	Data struct {
		ID     []byte   `bencode:"id"`
		Token  string   `bencode:"token"`
		Nodes  []byte   `bencode:"nodes"`
		Values []string `bencode:"values"`
	} `bencode:"r"`
}

func (p *KadResponse) ID() []byte {
	return p.Data.ID
}

func (p *KadResponse) Nodes() []byte {
	return p.Data.Nodes
}

func (p *KadResponse) Values() []string {
	return p.Data.Values
}

func EncodeCompactNode(nodes map[*ID]*net.UDPAddr) []byte {
	ip := make([]byte, 4)
	buf := bytes.NewBuffer(nil)
	for id, addr := range nodes {
		n, err := fmt.Sscanf(addr.IP.String(), "%d.%d.%d.%d", &ip[0], &ip[1], &ip[2], &ip[3])
		if err != nil || n != 4 {
			continue
		}
		buf.Write(id.Bytes())
		buf.Write(ip)
		buf.WriteByte(byte(addr.Port >> 8))
		buf.WriteByte(byte(addr.Port))
	}
	return buf.Bytes()
}

func DecodeCompactNode(b []byte) map[*ID]*net.UDPAddr {
	nodes := make(map[*ID]*net.UDPAddr)
	for i := 0; i < len(b)/26; i++ {
		bi := b[i*26:]
		id, err := NewID(bi[:20])
		if err != nil {
			continue
		}
		s := fmt.Sprintf("%d.%d.%d.%d:%d", bi[20], bi[21],
			bi[22], bi[23], (uint16(bi[24])<<8)|uint16(bi[25]))
		addr, err := net.ResolveUDPAddr("udp", s)
		if err != nil {
			continue
		}
		nodes[id] = addr
	}
	return nodes
}

func EncodeCompactPeer(peers []*Peer) [][]byte {
	return nil
}

func DecodeCompactPeer(b []string) []string {
	return nil
}
