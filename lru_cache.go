package main

// LRUCache for string->string
// TODO: Currently WIP, its just a map.
type LRUCache struct {
	m map[string]string
}

func CreateCache(capacity int) *LRUCache {
	return &LRUCache{m: make(map[string]string, capacity)}
}

func (l *LRUCache) Get(k string) (string, bool) {
	v, exists := l.m[k]
	return v, exists
}

func (l *LRUCache) Set(k, v string) {
	l.m[k] = v
}
