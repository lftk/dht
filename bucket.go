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
	first *ID
	time  time.Time
	nodes *list.List
}

// NewBucket return a bucket
func NewBucket(first *ID) *Bucket {
	return &Bucket{
		first: first,
		time:  time.Now(),
		nodes: list.New(),
	}
}

// Count returns count of all nodes
func (b *Bucket) Count() int {
	return b.nodes.Len()
}

// Append a node, move to back if exist node
func (b *Bucket) Append(n *Node) error {
	var updated bool
	b.handle(func(e *list.Element) bool {
		if e.Value.(*Node).id.Compare(n.id) == 0 {
			updated = true
			b.nodes.MoveToBack(e)
			return false
		}
		return true
	})
	if updated == false {
		if b.Count() == maxNodeCount {
			return errors.New("bucket is full")
		}
		b.nodes.PushBack(n)
	}
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
func (b *Bucket) Find(id *ID) *Node {
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
	var node *Node
	i := rand.Intn(b.Count())
	b.Map(func(n *Node) bool {
		if i--; i < 0 {
			node = n
			return false
		}
		return true
	})
	return node
}

// Map all node
func (b *Bucket) Map(f func(n *Node) bool) {
	b.handle(func(e *list.Element) bool {
		return f(e.Value.(*Node))
	})
}

func (b *Bucket) handle(f func(e *list.Element) bool) {
	for e := b.nodes.Front(); e != nil; e = e.Next() {
		if f(e) == false {
			return
		}
	}
}

func (b *Bucket) String() string {
	s := fmt.Sprintf("%v %d\n", b.first, b.Count())
	b.Map(func(n *Node) bool {
		s += fmt.Sprintf("  %v\n", n)
		return true
	})
	return s
}
