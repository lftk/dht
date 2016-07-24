package dht

import (
	"container/list"
	"fmt"
	"math/rand"
	"net"
	"time"
)

// Bucket type
type Bucket struct {
	cap   int
	first *ID
	time  time.Time
	nodes *list.List
}

// NewBucket return a bucket
func NewBucket(first *ID, cap int) *Bucket {
	return &Bucket{
		cap:   cap,
		first: first,
		time:  time.Now(),
		nodes: list.New(),
	}
}

// Count returns count of all nodes
func (b *Bucket) Count() int {
	return b.nodes.Len()
}

// Capacity returns bucket cap
func (b *Bucket) Capacity() int {
	return b.cap
}

// Insert a node, move to back if exist node
func (b *Bucket) Insert(id *ID, addr *net.UDPAddr) (n *Node) {
	b.handle(func(e *list.Element) bool {
		if e.Value.(*Node).id.Compare(id) == 0 {
			n = e.Value.(*Node)
			b.nodes.MoveToBack(e)
			return false
		}
		return true
	})
	if n == nil && b.Count() != b.cap {
		n = NewNode(id, addr)
		b.nodes.PushBack(n)
	}
	return
}

// Remove a node, return true if exist node
func (b *Bucket) Remove(id *ID) bool {
	for e := b.nodes.Front(); e != nil; e = e.Next() {
		if e.Value.(*Node).id.Compare(id) == 0 {
			b.nodes.Remove(e)
			return true
		}
	}
	return false
}

// Find returns node
func (b *Bucket) Find(id *ID) (node *Node) {
	b.Map(func(n *Node) bool {
		if n.id.Compare(id) == 0 {
			node = n
			return false
		}
		return true
	})
	return
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

// Time returns time
func (b *Bucket) Time() time.Time {
	return b.time
}

// Update time
func (b *Bucket) Update() {
	b.time = time.Now()
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

func (b *Bucket) clean(f func(n *Node) bool) {
	e := b.nodes.Front()
	for e != nil {
		next := e.Next()
		if f(e.Value.(*Node)) {
			b.nodes.Remove(e)
		}
		e = next
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
