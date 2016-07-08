package dht

import (
	"errors"
	"fmt"
	"sort"
)

// Table store all nodes
type Table struct {
	id      ID
	buckets *Bucket
}

// NewTable returns a table
func NewTable(id ID) *Table {
	t := &Table{
		id:      id,
		buckets: NewBucket(ZeroID),
	}
	return t
}

// Append a node
func (t *Table) Append(n *Node) error {
	if n.id.Compare(t.id) == 0 {
		return fmt.Errorf("node's id equal to table's id")
	}
	b := t.Find(n.id)
	if b == nil {
		return fmt.Errorf("not found bucket of %v", n.id)
	}
	if err := b.Append(n); err == nil {
		return nil
	}
	if b.Test(t.id) && t.split(b) {
		return t.Append(n)
	}
	return errors.New("drop this node")
}

func (t *Table) split(b *Bucket) bool {
	var bit int
	if b.next != nil {
		bit = b.next.first.LowBit()
	} else {
		bit = b.first.LowBit()
	}
	if bit++; bit >= 160 {
		return false
	}

	first, _ := NewID(b.first.Bytes())
	first.SetBit(bit, true)
	b2 := NewBucket(first)
	b2.next = b.next
	b.next = b2

	// switch to new bucket
	// ...

	return true
}

// Find returns bucket
func (t *Table) Find(id ID) (dst *Bucket) {
	t.Map(func(b *Bucket) bool {
		if b.Test(id) == true {
			dst = b
			return false
		}
		return true
	})
	return
}

// Lookup returns the K(8) closest good nodes
func (t *Table) Lookup(id ID) []*Node {
	b := t.Find(id)
	if b == nil {
		return nil
	}

	ln := newLookupNodes(id)
	if ln.CopyFrom(b); ln.Len() < 8 {
		prev, next := b.prev, b.next
		for ln.Len() < 8 && (prev != nil || next != nil) {
			if prev != nil {
				ln.CopyFrom(prev)
				prev = prev.prev
			}
			if next != nil {
				ln.CopyFrom(next)
				next = next.next
			}
		}
	}
	sort.Sort(ln)
	return ln.nodes[:8]
}

type lookupNodes struct {
	id    ID
	nodes []*Node
}

func newLookupNodes(id ID) *lookupNodes {
	return &lookupNodes{
		id:    id,
		nodes: make([]*Node, 0, 8),
	}
}

func (ln *lookupNodes) CopyFrom(b *Bucket) {
	b.Map(func(n *Node) bool {
		ln.nodes = append(ln.nodes, n)
		return true
	})
}

func (ln *lookupNodes) Len() int {
	return len(ln.nodes)
}

func (ln *lookupNodes) Less(i, j int) bool {
	for k := 0; k < 5; k++ {
		n1 := ln.nodes[i].id[k] ^ ln.id[k]
		n2 := ln.nodes[j].id[k] ^ ln.id[k]
		if n1 < n2 {
			return true
		} else if n1 > n2 {
			return false
		}
	}
	return true
}

func (ln *lookupNodes) Swap(i, j int) {
	ln.nodes[i], ln.nodes[j] = ln.nodes[j], ln.nodes[i]
}

// Map all buckets
func (t *Table) Map(f func(b *Bucket) bool) bool {
	for b := t.buckets; b != nil; b = b.next {
		if f(b) == false {
			return false
		}
	}
	return true
}

func (t *Table) String() string {
	var s string
	t.Map(func(b *Bucket) bool {
		s += fmt.Sprintf("%v\n", b)
		return true
	})
	return s
}
