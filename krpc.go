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

type kadQueryMessage struct {
	T []byte                 `bencode:"t"`
	Y string                 `bencode:"y"`
	Q string                 `bencode:"q"`
	A map[string]interface{} `bencode:"a"`
}

type kadReplyMessage struct {
	T []byte                 `bencode:"t"`
	Y string                 `bencode:"y"`
	R map[string]interface{} `bencode:"r"`
}

func newQueryMessage(tid []byte, q string, data map[string]interface{}) *kadQueryMessage {
	return &kadQueryMessage{tid, "q", q, data}
}

func newReplyMessage(tid []byte, data map[string]interface{}) *kadReplyMessage {
	return &kadReplyMessage{tid, "r", data}
}

type kadArguments struct {
	ID       []byte `bencode:"id"`
	Port     int64  `bencode:"port"`
	Token    []byte `bencode:"token"`
	Target   []byte `bencode:"target"`
	InfoHash []byte `bencode:"info_hash"`
}

type kadResponse struct {
	ID     []byte   `bencode:"id"`
	Token  []byte   `bencode:"token"`
	Nodes  []byte   `bencode:"nodes"`
	Values [][]byte `bencode:"values"`
}

type kadMessage struct {
	T []byte        `bencode:"t"`
	Y string        `bencode:"y"`
	Q string        `bencode:"q"`
	E []interface{} `bencode:"e"`
	A kadArguments  `bencode:"a"`
	R kadResponse   `bencode:"r"`
}

// ResolvePeer returns ip and port
func ResolvePeer(peer []byte) (ip string, port int) {
	if n := len(peer); n == 6 {
		ip = net.IPv4(peer[0], peer[1], peer[2], peer[3]).String()
		port = (int(peer[4]) << 8) | int(peer[5])
	}
	return
}

// ResolveNodes returns peers
func ResolveNodes(nodes []byte) (peers map[ID][]byte) {
	peers = make(map[ID][]byte)
	for i := 0; i < len(nodes)/26; i++ {
		node := nodes[i*26:]
		id, err := NewID(node[:20])
		if err == nil {
			peers[*id] = node[20:26]
		}
	}
	return
}
