package dht

import (
	"fmt"
	"net"

	"github.com/zeebo/bencode"
)

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
	Values [][]byte `bencode:"values"`
}

type KadMessage struct {
	T []byte        `bencode:"t"`
	Y string        `bencode:"y"`
	Q string        `bencode:"q"`
	E []interface{} `bencode:"e"`
	A KadArguments  `bencode:"a"`
	R KadResponse   `bencode:"r"`
}

func ResolveAddr(addr []byte) (ip string, port int) {
	if n := len(addr); n == 6 {
		ip = net.IPv4(addr[0], addr[1], addr[2], addr[3]).String()
		port = (int(addr[4]) << 8) | int(addr[5])
	}
	return
}
