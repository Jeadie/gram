package main

// ByteRing is a simple ring queue
type ByteRing struct {
	n    uint
	ring []byte
	i    uint
}

func CreateByteRing(capacity uint) ByteRing {
	return ByteRing{n: capacity, i: 0, ring: make([]byte, capacity)}
}

func (r *ByteRing) Insert(b byte) {
	r.ring[r.i] = b
	r.i = (r.i + 1) % r.n
}

func (r *ByteRing) GetHistory() []byte {
	hist := make([]byte, r.n)

	n := r.n

	for j := uint(0); j < n; j++ {
		hist[j] = r.ring[(r.i+(n-j-1))%n]
	}
	return hist
}
