package ratelimit_redis_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	redislib "github.com/webcore-go/lib-redis"
	"github.com/webcore-go/webcore/infra/config"
)

var podA, podB *redislib.Redis

func TestMain(m *testing.M) {
	if err := os.Chdir(testutil.RepoRoot()); err != nil {
		panic(err)
	}
	cfg := config.Config{}
	if err := config.LoadDefaultConfig(&cfg); err != nil {
		panic(err)
	}
	podA, podB = redislib.NewRedis(cfg.Redis), redislib.NewRedis(cfg.Redis)
	if err := podA.Client.Ping(context.Background()).Err(); err != nil {
		fmt.Fprintln(os.Stderr, "redis tidak hidup, test dilewati:", err)
		os.Exit(0)
	}
	code := m.Run()
	podA.Client.Close()
	podB.Client.Close()
	os.Exit(code)
}

func testKey(t *testing.T) string {
	k := "ratelimit:{uji-" + uuid.NewString() + "}"
	t.Cleanup(func() {
		iter := podA.Client.Scan(context.Background(), 0, k+"*", 100).Iterator()
		for iter.Next(context.Background()) {
			podA.Client.Del(context.Background(), iter.Val())
		}
	})
	return k
}

func TestKuotaDipakaiBersamaAntarPod(t *testing.T) {
	ctx := context.Background()
	key := testKey(t)

	var result []bool
	for i, pod := range []*redislib.Redis{podA, podB, podA, podB, podA} {
		h, err := pod.RateLimit(ctx, key, 4, time.Minute)
		require.NoError(t, err, "request ke-%d", i)
		result = append(result, h.Allowed)
	}
	assert.Equal(t, []bool{true, true, true, true, false}, result, "kuota 4 dihitung gabungan dua pod")

	h, err := podB.RateLimit(ctx, key, 4, time.Minute)
	require.NoError(t, err)
	assert.False(t, h.Allowed)
	assert.Zero(t, h.Remaining)
	assert.True(t, h.ResetAt.After(time.Now()))

	ttl, err := podA.Client.PTTL(ctx, fmt.Sprintf("%s:%d", key, h.ResetAt.UnixMilli()-time.Minute.Milliseconds())).Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Minute, "counter kedaluwarsa sendiri setelah dua window")
}

func TestSisaKuotaBerkurang(t *testing.T) {
	key := testKey(t)
	for _, remaining := range []int64{2, 1, 0} {
		h, err := podA.RateLimit(context.Background(), key, 3, time.Minute)
		require.NoError(t, err)
		assert.True(t, h.Allowed)
		assert.Equal(t, remaining, h.Remaining)
	}
}

func TestKuotaPulihSetelahWindowBerlalu(t *testing.T) {
	ctx := context.Background()
	key := testKey(t)
	for i := 0; i < 3; i++ {
		_, err := podA.RateLimit(ctx, key, 3, time.Second)
		require.NoError(t, err)
	}
	h, _ := podB.RateLimit(ctx, key, 3, time.Second)
	assert.False(t, h.Allowed)

	time.Sleep(2100 * time.Millisecond)
	h, err := podB.RateLimit(ctx, key, 3, time.Second)
	require.NoError(t, err)
	assert.True(t, h.Allowed)
}

func TestKonkurenDuaPodTepatSejumlahLimit(t *testing.T) {
	key := testKey(t)
	var lolos atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		pod := podA
		if i%2 == 1 {
			pod = podB
		}
		go func() {
			defer wg.Done()
			h, err := pod.RateLimit(context.Background(), key, 30, time.Minute)
			if err == nil && h.Allowed {
				lolos.Add(1)
			}
		}()
	}
	wg.Wait()
	assert.EqualValues(t, 30, lolos.Load())
}

func TestRedisMatiMengembalikanError(t *testing.T) {
	mati := redislib.NewRedis(config.RedisConfig{Host: "127.0.0.1", Port: 1})
	defer mati.Client.Close()
	ctx, batal := context.WithTimeout(context.Background(), time.Second)
	defer batal()
	_, err := mati.RateLimit(ctx, "ratelimit:{x}", 1, time.Minute)
	assert.Error(t, err)
}
