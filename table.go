package dht

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type table struct {
	id      hash
	buckets *list.List
}

func newTable(id hash) *table {
	t := &table{
		id:      id,
		buckets: list.New(),
	}
	min, _ := newHash(bytes.Repeat([]byte("00"), hashLen))
	max, _ := newHash(bytes.Repeat([]byte("ff"), hashLen))
	t.buckets.PushBack(newBucket(min, max))
	return t
}

func (t *table) append(n *node) error {
	if n.id.compare(t.id) == 0 {
		return fmt.Errorf("node's id equal to table's id")
	}

	var ele *list.Element
	for e := t.buckets.Front(); e != nil; e = e.Next() {
		if e.Value.(*bucket).contain(n.id) {
			ele = e
			break
		}
	}
	if ele == nil {
		return fmt.Errorf("not found bucket - %s", n.id)
	}

	b := ele.Value.(*bucket)
	if err := b.append(n); err == nil {
		return nil
	}
	if b.contain(t.id) == false {
		return errors.New("drop this node")
	}

	b1, b2 := b.split()
	b.clone(b1)
	b.cached = n
	t.buckets.InsertAfter(b2, ele)
	return errors.New("cache this node")
}

func (t *table) find(id hash) *bucket {
	for e := t.buckets.Front(); e != nil; e = e.Next() {
		if e.Value.(*bucket).contain(id) {
			return e.Value.(*bucket)
		}
	}
	return nil
}

func (t *table) String() string {
	var s string
	for e := t.buckets.Front(); e != nil; e = e.Next() {
		s += fmt.Sprintf("%v\n", e.Value.(*bucket))
	}
	return s
}

func init() {
	rand.Seed(time.Now().Unix())
}
