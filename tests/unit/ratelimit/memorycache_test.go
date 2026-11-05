package ratelimit_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	memlib "github.com/webcore-go/lib-memory"
	"github.com/webcore-go/webcore/infra/config"
)

func newCache(t *testing.T) *memlib.MemoryCache {
	t.Helper()
	c, err := memlib.NewMemoryCache(config.MemoryConfig{ExpiresIn: time.Minute})
	require.NoError(t, err)
	return c
}

func TestMemoryCacheKuotaDanSisa(t *testing.T) {
	c := newCache(t)
	for _, remaining := range []int64{2, 1, 0} {
		h, err := c.RateLimit(context.Background(), "a", 3, time.Minute)
		require.NoError(t, err)
		assert.True(t, h.Allowed)
		assert.Equal(t, remaining, h.Remaining)
	}
	h, err := c.RateLimit(context.Background(), "a", 3, time.Minute)
	require.NoError(t, err)
	assert.False(t, h.Allowed)
	assert.True(t, h.ResetAt.After(time.Now()))

	h, _ = c.RateLimit(context.Background(), "b", 3, time.Minute)
	assert.True(t, h.Allowed, "key lain punya kuota sendiri")
}

func TestMemoryCacheKuotaPulihSetelahWindowBerlalu(t *testing.T) {
	c := newCache(t)
	for i := 0; i < 3; i++ {
		c.RateLimit(context.Background(), "a", 3, time.Second)
	}
	h, _ := c.RateLimit(context.Background(), "a", 3, time.Second)
	assert.False(t, h.Allowed)

	time.Sleep(2100 * time.Millisecond)
	h, err := c.RateLimit(context.Background(), "a", 3, time.Second)
	require.NoError(t, err)
	assert.True(t, h.Allowed)
}

func TestMemoryCacheKonkurenTepatSejumlahLimit(t *testing.T) {
	c := newCache(t)
	var lolos atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h, err := c.RateLimit(context.Background(), "a", 50, time.Hour)
			if err == nil && h.Allowed {
				lolos.Add(1)
			}
		}()
	}
	wg.Wait()
	assert.EqualValues(t, 50, lolos.Load())
}

func TestMemoryCacheCounterKedaluwarsaSendiri(t *testing.T) {
	c := newCache(t)
	h, err := c.RateLimit(context.Background(), "a", 3, time.Hour)
	require.NoError(t, err)

	key := "a:" + itoa(h.ResetAt.Add(-time.Hour).UnixMilli())
	e, ok := c.Cache.GetEntryQuietly(key)
	require.True(t, ok)
	assert.InDelta(t, (2 * time.Hour).Seconds(), time.Until(e.ExpiresAt()).Seconds(), 5, "TTL dua window")
}

func TestMemoryCacheTTLTidakDipendekkanSaatDitolak(t *testing.T) {
	c := newCache(t)
	var reset time.Time
	for i := 0; i < 5; i++ {
		r, err := c.RateLimit(context.Background(), "a", 2, time.Hour)
		require.NoError(t, err)
		reset = r.ResetAt
	}
	key := "a:" + itoa(reset.Add(-time.Hour).UnixMilli())
	e, ok := c.Cache.GetEntryQuietly(key)
	require.True(t, ok)
	assert.Greater(t, time.Until(e.ExpiresAt()), time.Hour, "permintaan yang ditolak tidak memendekkan TTL ke memory.expires_in")
}
