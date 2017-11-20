package memtable

import (
	"math/bits"
	"math/rand"

	"github.com/golang/glog"
)

const (
	maxLevel = 16
)

type Memtable struct {
	head *node
	rnd  *rand.Rand
}

type node struct {
	key   string
	value []byte
	next  []*node
}

func New() *Memtable {
	h := &node{
		key:   "",
		value: nil,
		next:  make([]*node, maxLevel),
	}
	return &Memtable{
		head: h,
		rnd:  rand.New(rand.NewSource(134787)),
	}
}

// Insert inserts a key value pair into the memtable.
// Requires that the key does not already exist.
func (m *Memtable) Insert(key string, value []byte) {
	if key == "" {
		glog.Fatal("Invalid empty key.")
	}
	var prev [maxLevel]*node

	n := m.findGreaterOrEqual(key, prev[:])

	if n != nil && n.key == key {
		glog.Fatalf("Insert called with duplicate key %v.", key)
	}

	level := m.pickLevel()
	newNode := &node{
		key:   key,
		value: value,
		next:  make([]*node, level+1),
	}

	for i := 0; i <= level; i++ {
		newNode.next[i] = prev[i].next[i]
		prev[i].next[i] = newNode
	}
}

// findGreaterOrEqual retuns the first node that is greater than or equal to key.
// If prev is not nil, filled with the last node visited per level.
func (m *Memtable) findGreaterOrEqual(key string, prev []*node) *node {
	c := m.head
	cl := maxLevel - 1
	for {
		nextAtLevel := c.next[cl]
		if nextAtLevel != nil && nextAtLevel.key < key {
			c = nextAtLevel
		} else {
			if prev != nil {
				prev[cl] = c
			}
			if cl == 0 {
				return nextAtLevel
			}
			cl--
		}
	}
}

// Find returns value of key, or nil if not found.
func (m *Memtable) Find(key string) []byte {
	if key == "" {
		glog.Fatal("Invalid empty key.")
	}

	n := m.findGreaterOrEqual(key, nil)

	if n != nil && n.key == key {
		return n.value
	}
	return nil
}

// Iterator iterates entries in the memtable in ascending key order.
// Close() must be called after use.
type Iterator struct {
	m *Memtable
	n *node
}

// NewIterator creates an iterator for this memtable.
func (m *Memtable) NewIterator() *Iterator {
	return &Iterator{
		m: m,
		n: m.head,
	}
}

// Next advances the iterator. Returns true if there is a next value.
func (i *Iterator) Next() bool {
	i.n = i.n.next[0]
	return i.n != nil
}

// Key returns the current key.
func (i *Iterator) Key() string {
	return i.n.key
}

// Value returns the current value.
func (i *Iterator) Value() []byte {
	return i.n.value
}

// Close closes the iterator.
func (i *Iterator) Close() {}

// Level assigned to this node, zero indexed.
func (m *Memtable) pickLevel() int {
	var r uint64
	for r == 0 {
		r = uint64(m.rnd.Int63n(int64(1) << maxLevel))
	}
	return bits.TrailingZeros64(r)
}