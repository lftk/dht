package dht

import (
	"container/list"
	"errors"
	"fmt"
	"net"
	"sort"
)

// Table store all nodes
type Table struct {
	id      *ID
	buckets *list.List
}

// NewTable returns a table
func NewTable(id *ID) *Table {
	t := &Table{
		id:      id,
		buckets: list.New(),
	}
	t.buckets.PushBack(NewBucket(ZeroID))
	return t
}

// Append a node
func (t *Table) Append(n *Node) error {
	if n.id.Compare(t.id) == 0 {
		return fmt.Errorf("node's id equal to table's id")
	}
	e := t.find(n.id)
	if e == nil {
		return fmt.Errorf("not found bucket of %v", n.id)
	}
	b := e.Value.(*Bucket)
	if err := b.Append(n); err == nil {
		return nil
	}
	if t.selfInBucket(e) && t.splitBucket(e) {
		return t.Append(n)
	}
	return errors.New("drop this node")
}

func (t *Table) Insert(id *ID, addr *net.UDPAddr) (*Node, error) {
	if id.Compare(t.id) == 0 {
		return nil, errors.New("id equal to table's id")
	}
	return t.insert(id, addr)
}

func (t *Table) insert(id *ID, addr *net.UDPAddr) (n *Node, err error) {
	if e := t.find(id); e != nil {
		if n = e.Value.(*Bucket).Insert(id, addr); n != nil {
			return
		}
		if t.selfInBucket(e) && t.splitBucket(e) {
			return t.insert(id, addr)
		}
	}
	err = errors.New("drop this node")
	return
}

func (t *Table) selfInBucket(e *list.Element) bool {
	return inBucket(t.id, e)
}

func (t *Table) splitBucket(e *list.Element) bool {
	bit := e.Value.(*Bucket).first.LowBit()
	if e.Next() != nil {
		bit2 := e.Next().Value.(*Bucket).first.LowBit()
		if bit < bit2 {
			bit = bit2
		}
	}
	if bit++; bit >= 160 {
		return false
	}

	b := e.Value.(*Bucket)
	first, _ := NewID(b.first.Bytes())
	first.SetBit(bit, true)
	b2 := NewBucket(first)
	t.buckets.InsertAfter(b2, e)

	var eles []*list.Element
	b.handle(func(be *list.Element) bool {
		if inBucket(be.Value.(*Node).id, e) == false {
			eles = append(eles, be)
		}
		return true
	})
	for _, ele := range eles {
		b2.nodes.PushBack(b.nodes.Remove(ele))
	}

	return true
}

// Find returns bucket
func (t *Table) Find(id *ID) *Bucket {
	if e := t.find(id); e != nil {
		return e.Value.(*Bucket)
	}
	return nil
}

func (t *Table) find(id *ID) (ele *list.Element) {
	t.handle(func(e *list.Element) bool {
		if inBucket(id, e) {
			ele = e
			return false
		}
		return true
	})
	return
}

// Lookup returns the K(8) closest good nodes
func (t *Table) Lookup(id *ID) []*Node {
	e := t.find(id)
	if e == nil {
		return nil
	}

	ln := newLookupNodes(id)
	if ln.CopyFrom(e); ln.Len() < maxNodeCount {
		prev, next := e.Prev(), e.Next()
		for ln.Len() < 8 && (prev != nil || next != nil) {
			if prev != nil {
				ln.CopyFrom(prev)
				prev = prev.Prev()
			}
			if next != nil {
				ln.CopyFrom(next)
				next = next.Next()
			}
		}
	}
	sort.Sort(ln)

	if ln.Len() > maxNodeCount {
		return ln.nodes[:maxNodeCount]
	}
	return ln.nodes
}

func inBucket(id *ID, e *list.Element) bool {
	if b := e.Value.(*Bucket); b.first.Compare(id) > 0 {
		return false
	}
	if n := e.Next(); n != nil {
		if b := n.Value.(*Bucket); b.first.Compare(id) <= 0 {
			return false
		}
	}
	return true
}

type lookupNodes struct {
	id    *ID
	nodes []*Node
}

func newLookupNodes(id *ID) *lookupNodes {
	return &lookupNodes{
		id:    id,
		nodes: make([]*Node, 0, maxNodeCount),
	}
}

func (ln *lookupNodes) CopyFrom(e *list.Element) {
	e.Value.(*Bucket).Map(func(n *Node) bool {
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
func (t *Table) Map(f func(b *Bucket) bool) {
	t.handle(func(e *list.Element) bool {
		return f(e.Value.(*Bucket))
	})
}

func (t *Table) handle(f func(e *list.Element) bool) {
	for e := t.buckets.Front(); e != nil; e = e.Next() {
		if f(e) == false {
			return
		}
	}
}

func (t *Table) String() (s string) {
	var n int
	t.Map(func(b *Bucket) bool {
		n += b.Count()
		s += fmt.Sprintf("%v\n", b)
		return true
	})
	s = fmt.Sprintf("%v %d\n\n", t.id, n) + s
	return
}
