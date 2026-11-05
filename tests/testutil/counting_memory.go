package testutil

import (
	"sync/atomic"

	"github.com/webcore-go/webcore/port"
)

// CountingMemory menghitung Get ke cache; di produksi tiap Get satu round-trip Redis.
type CountingMemory struct {
	port.ICacheMemory
	gets, hits atomic.Int64
}

func (m *CountingMemory) Get(key string, outvalue any) bool {
	m.gets.Add(1)
	ok := m.ICacheMemory.Get(key, outvalue)
	if ok {
		m.hits.Add(1)
	}
	return ok
}

// CacheGets mengembalikan jumlah Get dan yang ketemu sejak awal.
func (m *CountingMemory) CacheGets() (gets, hits int64) { return m.gets.Load(), m.hits.Load() }
