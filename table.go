package dht

import (
	"container/list"
	"errors"
	"fmt"
)

type Table struct {
	id      ID
	buckets *list.List
}

// NewTable returns a table
func NewTable(id ID) *Table {
	t := &Table{
		id:      id,
		buckets: list.New(),
	}
	// initialize the buckets
	for i := 0; i < 32; i++ {
		b := NewBucket(ZeroID, 155)
		b.first[0] = uint32(i) << (32 - 5)
		t.buckets.PushBack(b)
	}
	return t
}

// Append a node
func (t *Table) Append(n *Node) error {
	if n.id.Compare(t.id) == 0 {
		return fmt.Errorf("node's id equal to table's id")
	}

	var ele *list.Element
	for e := t.buckets.Front(); e != nil; e = e.Next() {
		if e.Value.(*Bucket).Test(n.id) {
			ele = e
			break
		}
	}
	if ele == nil {
		return fmt.Errorf("not found bucket - %s", n.id)
	}

	b := ele.Value.(*Bucket)
	if err := b.Append(n); err == nil {
		return nil
	}
	if b.Test(t.id) && t.split(ele) {
		return t.Append(n)
	}
	return errors.New("drop this node")
}

func (t *Table) split(e *list.Element) bool {
	b1 := e.Value.(*Bucket)
	if b1.span == 0 {
		return false
	}

	b1.span--
	first := b1.first
	slot := 4 - (b1.span >> 5)
	mask := uint32(1 << (b1.span & 31))
	b2 := NewBucket(first, b1.span)
	b2.first[slot] |= mask
	t.buckets.InsertAfter(b2, e)

	// switch to new bucket
	ele := b1.nodes.Front()
	for ele != nil {
		next := ele.Next()
		if n := ele.Value.(*Node); n.id[slot]&mask != 0 {
			b1.nodes.Remove(ele)
			b2.nodes.PushBack(n)
		}
		ele = next
	}
	return true
}

/*
func (t *table) Find(id ID) *bucket {
	for e := t.buckets.Front(); e != nil; e = e.Next() {
		if e.Value.(*bucket).contain(id) {
			return e.Value.(*bucket)
		}
	}
	return nil
}
*/

func (t *Table) String() string {
	var s string
	for e := t.buckets.Front(); e != nil; e = e.Next() {
		s += fmt.Sprintf("%v\n", e.Value.(*Bucket))
	}
	return s
}
