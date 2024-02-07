package skiplist

import (
	"errors"
	"github.com/liyue201/gostl/utils/comparator"
	"github.com/liyue201/gostl/utils/sync"
	"github.com/liyue201/gostl/utils/visitor"
	"math/rand"
	gosync "sync"
	"time"
)

var (
	defaultMaxLevel = 10
	defaultLocker   sync.FakeLocker
)
var ErrorNotFound = errors.New("not found")

// Options holds Skiplist's options
type Options struct {
	maxLevel int
	locker   sync.Locker
}

// Option is a function used to set Options
type Option func(option *Options)

// WithGoroutineSafe sets Skiplist goroutine-safety,
func WithGoroutineSafe() Option {
	return func(option *Options) {
		option.locker = &gosync.RWMutex{}
	}
}

// WithMaxLevel sets max level of Skiplist
func WithMaxLevel(maxLevel int) Option {
	return func(option *Options) {
		option.maxLevel = maxLevel
	}
}

// Node is a list node
type Node[K, V any] struct {
	Next []*Element[K, V]
}

// Element is a kind of node with Key-Value data
type Element[K, V any] struct {
	Node[K, V]
	Key   K
	Value V
}

// Skiplist is a kind of data structure which can search quickly by exchanging space for time
type Skiplist[K, V any] struct {
	locker         sync.Locker
	Head           Node[K, V]
	maxLevel       int
	keyCmp         comparator.Comparator[K]
	len            int
	prevNodesCache []*Node[K, V]
	rander         *rand.Rand
}

// New news a Skiplist
func New[K, V any](cmp comparator.Comparator[K], opts ...Option) *Skiplist[K, V] {
	option := Options{
		maxLevel: defaultMaxLevel,
		locker:   defaultLocker,
	}
	for _, opt := range opts {
		opt(&option)
	}
	l := &Skiplist[K, V]{
		locker:   option.locker,
		maxLevel: option.maxLevel,
		keyCmp:   cmp,
		rander:   rand.New(rand.NewSource(time.Now().Unix())),
	}
	l.Head.Next = make([]*Element[K, V], l.maxLevel)
	l.prevNodesCache = make([]*Node[K, V], l.maxLevel)
	return l
}

// Insert inserts a Key-Value pair into the skiplist
func (sl *Skiplist[K, V]) Insert(key K, value V) {
	sl.locker.Lock()
	defer sl.locker.Unlock()
	prevs := sl.findPrevNodes(key)

	if prevs[0].Next[0] != nil && sl.keyCmp(prevs[0].Next[0].Key, key) == 0 {
		//same Key, update Value
		prevs[0].Next[0].Value = value
		return
	}

	level := sl.randomLevel()

	e := &Element[K, V]{
		Key:   key,
		Value: value,
		Node: Node[K, V]{
			Next: make([]*Element[K, V], level),
		},
	}

	for i := range e.Next {
		e.Next[i] = prevs[i].Next[i]
		prevs[i].Next[i] = e
	}

	sl.len++
}

// Get returns the Value associated with the passed Key if the Key is in the skiplist, otherwise returns error
func (sl *Skiplist[K, V]) Get(key K) (V, error) {
	var pre = &sl.Head
	for i := sl.maxLevel - 1; i >= 0; i-- {
		cur := pre.Next[i]
		for ; cur != nil; cur = cur.Next[i] {
			cmpRet := sl.keyCmp(cur.Key, key)
			if cmpRet == 0 {
				return cur.Value, nil
			}
			if cmpRet > 0 {
				break
			}
			pre = &cur.Node
		}
	}
	return *new(V), ErrorNotFound
}

// Remove removes the Key-Value pair associated with the passed Key and returns true if the Key is in the skiplist, otherwise returns false
func (sl *Skiplist[K, V]) Remove(key K) bool {
	sl.locker.Lock()
	defer sl.locker.Unlock()

	prevs := sl.findPrevNodes(key)
	element := prevs[0].Next[0]
	if element == nil {
		return false
	}
	if element != nil && sl.keyCmp(element.Key, key) != 0 {
		return false
	}

	for i, v := range element.Next {
		prevs[i].Next[i] = v
	}
	sl.len--
	return true
}

// Len returns the amount of Key-Value pair in the skiplist
func (sl *Skiplist[K, V]) Len() int {
	sl.locker.RLock()
	defer sl.locker.RUnlock()
	return sl.len
}

func (sl *Skiplist[K, V]) randomLevel() int {
	total := uint64(1)<<uint64(sl.maxLevel) - 1 // 2^n-1
	k := sl.rander.Uint64() % total
	levelN := uint64(1) << (uint64(sl.maxLevel) - 1)

	level := 1
	for total -= levelN; total > k; level++ {
		levelN >>= 1
		total -= levelN
	}
	return level
}

func (sl *Skiplist[K, V]) findPrevNodes(key K) []*Node[K, V] {
	prevs := sl.prevNodesCache
	prev := &sl.Head
	for i := sl.maxLevel - 1; i >= 0; i-- {
		if sl.Head.Next[i] != nil {
			for next := prev.Next[i]; next != nil; next = next.Next[i] {
				if sl.keyCmp(next.Key, key) >= 0 {
					break
				}
				prev = &next.Node
			}
		}
		prevs[i] = prev
	}
	return prevs
}

// Traversal traversals elements in the skiplist, it will stop until to the end or the visitor returns false
func (sl *Skiplist[K, V]) Traversal(visitor visitor.KvVisitor[K, V]) {
	for e := sl.Head.Next[0]; e != nil; e = e.Next[0] {
		if !visitor(e.Key, e.Value) {
			return
		}
	}
}

// Keys returns all keys in the skiplist
func (sl *Skiplist[K, V]) Keys() []K {
	var keys []K
	sl.Traversal(func(key K, value V) bool {
		keys = append(keys, key)
		return true
	})
	return keys
}
