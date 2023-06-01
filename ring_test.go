package main

import (
	"reflect"
	"testing"
)

func TestByteRing(t *testing.T) {
	r := NewbyteRing(3)

	r.Insert('a')
	r.Insert('b')
	r.Insert('c')

	if got := r.GetHistory(); !reflect.DeepEqual(got, []byte{'c', 'b', 'a'}) {
		t.Errorf("GetHistory() = %v, want %v", got, []byte{'c', 'b', 'a'})
	}

	r.Insert('d')

	if got := r.GetHistory(); !reflect.DeepEqual(got, []byte{'d', 'c', 'b'}) {
		t.Errorf("GetHistory() = %v, want %v", got, []byte{'d', 'c', 'b'})
	}

	r.Insert('e')
	r.Insert('f')

	if got := r.GetHistory(); !reflect.DeepEqual(got, []byte{'f', 'e', 'd'}) {
		t.Errorf("GetHistory() = %v, want %v", got, []byte{'f', 'e', 'd'})
	}
}
