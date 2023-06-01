package main

// byteRing is a simple ring queue
type byteRing struct {
	n    uint   // capacity of the ring buffer
	ring []byte // the ring buffer itself
	i    uint   // current index in the ring buffer
}

// NewbyteRing creates a new byteRing with a specified capacity.
func NewbyteRing(capacity uint) *byteRing {
	return &byteRing{n: capacity, i: 0, ring: make([]byte, capacity)}
}

// Insert inserts a byte into the ring buffer at the current index.
// If the buffer is full, it overwrites the oldest entry.
func (r *byteRing) Insert(b byte) {
	r.ring[r.i] = b
	r.i = (r.i + 1) % r.n
}

// GetHistory returns a copy of the current state of the ring buffer,
// with the most recently inserted byte first.
func (r *byteRing) GetHistory() []byte {
	hist := make([]byte, r.n)
	for j := uint(0); j < r.n; j++ {
		hist[j] = r.ring[(r.i+r.n-j-1)%r.n]
	}
	return hist
}
