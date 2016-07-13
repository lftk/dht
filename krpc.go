package dht

import (
	"fmt"

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

type KadErrorMessage struct {
	Data []interface{} `bencode:"e"`
}

func (e *KadErrorMessage) Value() int64 {
	return e.Data[0].(int64)
}

func (e *KadErrorMessage) String() string {
	return e.Data[1].(string)
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
