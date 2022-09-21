package main

// LRUCache for string->string
// TODO: Currently WIP, its just a map.
type LRUCache struct {
	m map[string]LinkedNode
	n int

	lru *LinkedNode
	mru *LinkedNode
}

func CreateCache(capacity int) *LRUCache {
	return &LRUCache{
		m: make(map[string]LinkedNode, capacity),
		n: capacity,
	}
}

func (l *LRUCache) Get(k string) (string, bool) {
	v, exists := l.pop(k)
	if !exists {
		return "", false
	}

	l.Set(k, v)

	return v, exists
}

func (l *LRUCache) pop(k string) (string, bool) {
	node, exists := l.m[k]
	if !exists {
		return "", false
	}

	node.Remove()

	return node.v, true
}

func (l *LRUCache) Set(k, v string) {
	if len(l.m)+1 > l.n {
		l.evictLRU()
	}

	l.setNewMru(k, v)
	l.m[k] = *l.mru
}

func (l *LRUCache) setNewMru(k, v string) {
	n := &LinkedNode{
		k: k,
		v: v,
	}
	if l.mru != nil {
		currMru := l.mru
		n.prev = currMru
		currMru.next = n

	}
	l.mru = n
}

func (l *LRUCache) evictLRU() {
	currLru := l.lru
	if currLru != nil {
		l.lru = currLru.next
		delete(l.m, currLru.k)
	}
}

// LinkedNode structure for the basis of a doubly-linked list.
type LinkedNode struct {
	prev *LinkedNode
	next *LinkedNode
	k, v string
}

// Remove a LinkedNode from its neighbours.
func (n *LinkedNode) Remove() {
	if n.prev != nil {
		n.prev.next = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	}
}
