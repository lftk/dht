package dht

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ZeroID "0000000000000000000000000000000000000000"
var ZeroID = new(ID)

// ID consists of 160 bits
type ID [20]byte

// NewID returns a id
func NewID(b []byte) (*ID, error) {
	if len(b) != 20 {
		return nil, fmt.Errorf("invalid hash")
	}
	id := new(ID)
	for i := 0; i < 20; i++ {
		id[i] = b[i]
	}
	return id, nil
}

// ResolveID returns a id
func ResolveID(s string) (*ID, error) {
	id := new(ID)
	n, err := hex.Decode(id[:], []byte(s))
	if err != nil {
		return nil, err
	}
	if n != 20 {
		return nil, fmt.Errorf("invalid hash")
	}
	return id, nil
}

var idRand = rand.New(rand.NewSource(time.Now().UnixNano()))
var idLock = new(sync.RWMutex)

// NewRandomID returns a random id
func NewRandomID() *ID {
	//	idLock.Lock()
	//	defer idLock.Unlock()

	id := new(ID)
	n, err := idRand.Read(id[:])
	if err != nil || n != 20 {
		return ZeroID
	}
	return id
}

// Compare two id
func (id *ID) Compare(o *ID) int {
	for i := 0; i < 20; i++ {
		if id[i] < o[i] {
			return -1
		} else if id[i] > o[i] {
			return 1
		}
	}
	return 0
}

// LowBit find the lowest 1 bit in an id
func (id *ID) LowBit() int {
	var i, j int
	for i = 19; i >= 0; i-- {
		if id[i] != 0 {
			break
		}
	}
	if i < 0 {
		return -1
	}
	for j = 7; j >= 0; j-- {
		if (id[i] & (0x80 >> uint32(j))) != 0 {
			break
		}
	}
	return 8*i + j
}

// SetBit set bit
func (id *ID) SetBit(i int, b bool) error {
	if b {
		id[i/8] |= (0x80 >> uint32(i%8))
	} else {
		id[i/8] &= ^(0x80 >> uint32(i%8))
	}
	return nil
}

// GetBit return bit
func (id *ID) GetBit(i int) (bool, error) {
	n := id[i/8] & (0x80 >> uint32(i%8))
	return n != 0, nil
}

// Bytes return 20 bytes
func (id *ID) Bytes() []byte {
	return id[:]
}

func (id *ID) String() string {
	return hex.EncodeToString(id[:])
}
