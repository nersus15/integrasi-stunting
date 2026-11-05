package ratelimit_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webcore-go/webcore/infra/middleware"
)

type clock struct{ t time.Time }

func (j *clock) now() time.Time          { return j.t }
func (j *clock) advance(d time.Duration) { j.t = j.t.Add(d) }
func storeWithClock() (*middleware.MemoryRateLimitStore, *clock) {
	j := &clock{t: time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)}
	s := middleware.NewMemoryRateLimitStore()
	s.Now = j.now
	return s, j
}

func fetch(t *testing.T, s *middleware.MemoryRateLimitStore, key string) (bool, int64) {
	t.Helper()
	h, err := s.RateLimit(context.Background(), key, 3, time.Minute)
	require.NoError(t, err)
	return h.Allowed, h.Remaining
}

func TestMemoriKuotaPerKey(t *testing.T) {
	s, j := storeWithClock()

	for _, remaining := range []int64{2, 1, 0} {
		ok, r := fetch(t, s, "a")
		assert.True(t, ok)
		assert.Equal(t, remaining, r)
	}
	ok, r := fetch(t, s, "a")
	assert.False(t, ok, "melewati kuota")
	assert.Zero(t, r)

	ok, _ = fetch(t, s, "b")
	assert.True(t, ok, "key lain punya kuota sendiri")

	h, _ := s.RateLimit(context.Background(), "a", 3, time.Minute)
	assert.True(t, j.t.Add(time.Minute).Equal(h.ResetAt), "reset di akhir window berjalan: %v", h.ResetAt)
}

func TestMemoriSlidingWindow(t *testing.T) {
	s, j := storeWithClock()
	for i := 0; i < 3; i++ {
		fetch(t, s, "a")
	}

	j.advance(time.Minute)
	ok, _ := fetch(t, s, "a")
	assert.False(t, ok, "awal window baru: window lalu masih dihitung penuh, tidak ada lonjakan 2x")

	j.advance(30 * time.Second)
	ok, _ = fetch(t, s, "a")
	assert.True(t, ok, "setengah window: bobot window lalu tinggal 1.5")
	ok, _ = fetch(t, s, "a")
	assert.False(t, ok)

	j.advance(2 * time.Minute)
	for i := 0; i < 3; i++ {
		ok, _ = fetch(t, s, "a")
		assert.True(t, ok, "setelah dua window kuota kembali penuh")
	}
}

func TestMemoriCleanup(t *testing.T) {
	s, j := storeWithClock()
	fetch(t, s, "lama")
	j.advance(time.Minute)
	fetch(t, s, "baru")

	s.Cleanup()
	assert.Equal(t, 2, s.Len(), "yang baru lewat satu window masih dibutuhkan untuk bobot")

	j.advance(time.Minute)
	s.Cleanup()
	assert.Equal(t, 1, s.Len())
	j.advance(time.Minute)
	s.Cleanup()
	assert.Zero(t, s.Len())
}

func TestMemoriStartCleanupBerhentiSaatContextSelesai(t *testing.T) {
	s := middleware.NewMemoryRateLimitStore()
	waktu := time.Now().Add(-time.Hour)
	s.Now = func() time.Time { return waktu }
	_, err := s.RateLimit(context.Background(), "a", 1, time.Millisecond)
	require.NoError(t, err)
	s.Now = nil

	ctx, batal := context.WithCancel(context.Background())
	s.StartCleanup(ctx, 5*time.Millisecond)
	require.Eventually(t, func() bool { return s.Len() == 0 }, time.Second, 5*time.Millisecond)
	batal()
}

func TestMemoriKonkurenTepatSejumlahLimit(t *testing.T) {
	s := middleware.NewMemoryRateLimitStore()
	var lolos atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h, err := s.RateLimit(context.Background(), "a", 50, time.Hour)
			if err == nil && h.Allowed {
				lolos.Add(1)
			}
		}()
	}
	wg.Wait()
	assert.EqualValues(t, 50, lolos.Load())
}
