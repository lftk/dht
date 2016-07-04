package dht

import (
	"container/list"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

const maxNodeCount int = 8

// Bucket type
type Bucket struct {
	first ID
	span  uint
	time  int64
	nodes *list.List
}

// NewBucket return a bucket
func NewBucket(first ID, span uint) *Bucket {
	return &Bucket{
		first: first,
		span:  span,
		time:  time.Now().Unix(),
		nodes: list.New(),
	}
}

// Count returns count of all node
func (b *Bucket) Count() int {
	return b.nodes.Len()
}

// Test returns true if has same prefix
func (b *Bucket) Test(id ID) bool {
	for i, span := 0, 160-int(b.span); span > 0; i, span = i+1, span-32 {
		mask := uint32(0xFFFFFFFF)
		if 32-span > 0 {
			mask <<= uint32(32 - span)
		}
		if (b.first[i]^id[i])&mask != 0 {
			return false
		}
	}
	return true
}

// Append a node, move to back if exist node
func (b *Bucket) Append(n *Node) error {
	for e := b.nodes.Front(); e != nil; e = e.Next() {
		if e.Value.(*Node).id.Compare(n.id) == 0 {
			b.nodes.MoveToBack(e)
			return nil
		}
	}
	if b.Count() == maxNodeCount {
		return errors.New("bucket is full")
	}
	b.nodes.PushBack(n)
	return nil
}

// Remove a node, return true if exist node
func (b *Bucket) Remove(n *Node) bool {
	for e := b.nodes.Front(); e != nil; e = e.Next() {
		if e.Value == n {
			b.nodes.Remove(e)
			return true
		}
	}
	return false
}

// Find returns node
func (b *Bucket) Find(id ID) *Node {
	var ptr *Node
	b.Map(func(n *Node) bool {
		if n.id.Compare(id) == 0 {
			ptr = n
			return false
		}
		return true
	})
	return ptr
}

// Random returns a random node
func (b *Bucket) Random() *Node {
	if b.Count() == 0 {
		return nil
	}
	var ptr *Node
	i := rand.Intn(b.Count())
	b.Map(func(n *Node) bool {
		if i--; i < 0 {
			ptr = n
			return false
		}
		return true
	})
	return ptr
}

// Map all node
func (b *Bucket) Map(f func(n *Node) bool) bool {
	for e := b.nodes.Front(); e != nil; e = e.Next() {
		if f(e.Value.(*Node)) == false {
			return false
		}
	}
	return true
}

func (b *Bucket) String() string {
	s := fmt.Sprintf("%v %d %d\n", b.first, b.span, b.Count())
	for e := b.nodes.Front(); e != nil; e = e.Next() {
		s += fmt.Sprintf("  %v\n", e.Value.(*Node))
	}
	return s
}
