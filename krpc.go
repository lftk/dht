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

type KadArguments struct {
	ID       []byte `bencode:"id"`
	Port     int64  `bencode:"port"`
	Token    []byte `bencode:"token"`
	Target   []byte `bencode:"target"`
	InfoHash []byte `bencode:"info_hash"`
}

type KadResponse struct {
	ID     []byte   `bencode:"id"`
	Token  []byte   `bencode:"token"`
	Nodes  []byte   `bencode:"nodes"`
	Values []string `bencode:"values"`
}

type KadMessage struct {
	T []byte        `bencode:"t"`
	Y string        `bencode:"y"`
	Q string        `bencode:"q"`
	E []interface{} `bencode:"e"`
	A KadArguments  `bencode:"a"`
	R KadResponse   `bencode:"r"`
}
