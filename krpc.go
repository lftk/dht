package dht

import (
	"fmt"

	"github.com/zeebo/bencode"
)

// krpc protocol message packet
type packet struct {
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

func (p *packet) Error() (n int64, s string) {
	if len(p.E) == 2 {
		n = p.E[0].(int64)
		s = p.E[1].(string)
	}
	return
}

func (p *packet) Marshal() ([]byte, error) {
	return bencode.EncodeBytes(p)
}

func (p *packet) Unmarshal(b []byte) (err error) {
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("panic() when bencode.DecodeBytes")
		}
	}()
	err = bencode.DecodeBytes(b, p)
	return
}
