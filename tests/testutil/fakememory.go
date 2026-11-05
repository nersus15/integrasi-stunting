package testutil

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/webcore-go/webcore/port"
)

// FakeMemory adalah implementasi in-memory sederhana dari port.ICacheMemory.
type FakeMemory struct {
	mu    sync.RWMutex
	items map[string][]byte
}

var _ port.ICacheMemory = (*FakeMemory)(nil)

func NewFakeMemory() *FakeMemory { return &FakeMemory{items: make(map[string][]byte)} }

func (m *FakeMemory) Install(args ...any) error { return nil }
func (m *FakeMemory) Uninstall() error          { return nil }
func (m *FakeMemory) Connect() error            { return nil }

func (m *FakeMemory) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = make(map[string][]byte)
	return nil
}

func (m *FakeMemory) Set(key string, value any, ttl time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = raw
	return nil
}

func (m *FakeMemory) Get(key string, outvalue any) bool {
	m.mu.RLock()
	raw, ok := m.items[key]
	m.mu.RUnlock()

	if !ok {
		return false
	}
	return json.Unmarshal(raw, outvalue) == nil
}

func (m *FakeMemory) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.items[key]; !ok {
		return fmt.Errorf("Gagal delete key: %s, tidak ditemukan", key)
	}

	delete(m.items, key)
	return nil
}

func (m *FakeMemory) Keys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.items))
	for k := range m.items {
		keys = append(keys, k)
	}
	return keys
}
