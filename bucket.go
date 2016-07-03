package dht

import (
	"container/list"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

const maxNodeCount int = 8

type bucket struct {
	min    hash
	max    hash
	time   int64
	nodes  *list.List
	cached *node
}

func newBucket(min, max hash) *bucket {
	return &bucket{
		min:   min,
		max:   max,
		time:  time.Now().Unix(),
		nodes: list.New(),
	}
}

func (b *bucket) count() int {
	return b.nodes.Len()
}

func (b *bucket) clone(o *bucket) {
	if b != o {
		b.min = o.min
		b.max = o.max
		b.time = o.time
		b.nodes = o.nodes
		b.cached = o.cached
	}
}

func (b *bucket) contain(id hash) bool {
	return id.inRange(b.min, b.max)
}

func (b *bucket) append(n *node) error {
	for e := b.nodes.Front(); e != nil; e = e.Next() {
		if e.Value.(*node).id.compare(n.id) == 0 {
			b.nodes.MoveToBack(e)
			return nil
		}
	}
	if b.count() == maxNodeCount {
		return errors.New("bucket is full")
	}
	b.nodes.PushBack(n)
	return nil
}

func (b *bucket) remove(n *node) bool {
	for e := b.nodes.Front(); e != nil; e = e.Next() {
		if e.Value == n {
			b.nodes.Remove(e)
			return true
		}
	}
	return false
}

func (b *bucket) find(id hash) *node {
	var ptr *node
	b.handle(func(n *node) bool {
		if n.id.compare(id) == 0 {
			ptr = n
			return false
		}
		return true
	})
	return ptr
}

func (b *bucket) random() *node {
	if b.count() == 0 {
		return nil
	}
	var ptr *node
	i := rand.Intn(b.count())
	b.handle(func(n *node) bool {
		if i--; i < 0 {
			ptr = n
			return false
		}
		return true
	})
	return ptr
}

func (b *bucket) split() (b1, b2 *bucket) {
	if b.cached != nil {
		b.cached.ping()
		b.cached = nil
	}

	mid := b.min.middle(b.max)
	b1 = newBucket(b.min, mid)
	b2 = newBucket(mid, b.max)
	b.handle(func(n *node) bool {
		if n.id.compare(mid) < 0 {
			b1.nodes.PushBack(n)
		} else {
			b2.nodes.PushBack(n)
		}
		return true
	})
	return
}

func (b *bucket) handle(f func(n *node) bool) bool {
	for e := b.nodes.Front(); e != nil; e = e.Next() {
		if f(e.Value.(*node)) == false {
			return false
		}
	}
	return true
}

func (b *bucket) String() string {
	s := fmt.Sprintf("%v-%v %d\n", b.min, b.max, b.count())
	for e := b.nodes.Front(); e != nil; e = e.Next() {
		s += fmt.Sprintf("  %v\n", e.Value.(*node))
	}
	return s
}
