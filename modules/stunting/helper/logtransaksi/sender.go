package logtransaksi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/webcore-go/webcore/infra/logger"
)

type Sender struct {
	*Masker
	cfg    config.LogTransaksiConfig
	rdb    *redis.Client
	client *http.Client

	ctx    context.Context
	cancel context.CancelFunc
	stop   chan struct{}
	wg     sync.WaitGroup
	once   sync.Once

	down   atomic.Bool
	failed atomic.Int64
	lost   atomic.Int64
}

const popTimeout = time.Second

func New(cfg config.LogTransaksiConfig, rdb *redis.Client) (*Sender, error) {
	if rdb == nil {
		return nil, errors.New("log transaksi butuh Redis sebagai antrean")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.RetryInterval <= 0 {
		cfg.RetryInterval = 10 * time.Second
	}
	if cfg.RedisKey == "" {
		cfg.RedisKey = config.DefaultLogRedisKey
	}
	if cfg.Mask == nil {
		cfg.Mask = config.DefaultLogMask
	}
	if cfg.Workers <= 0 {
		cfg.Workers = config.DefaultLogWorkers
	}
	headers, err := mergeHeaders(cfg.Headers, cfg.RawHeaders)
	if err != nil {
		return nil, err
	}
	cfg.Headers = headers
	logger.Info("LogTransaksi: ", "cfg", cfg)
	rctx, rcancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer rcancel()
	if err := rdb.Ping(rctx).Err(); err != nil {
		return nil, fmt.Errorf("redis log transaksi: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &Sender{
		Masker: NewMasker(cfg.Mask),
		cfg:    cfg,
		rdb:    rdb,
		client: &http.Client{Timeout: cfg.Timeout},
		ctx:    ctx,
		cancel: cancel,
		stop:   make(chan struct{}),
	}

	jobs := make(chan []byte)
	ready := make(chan struct{}, cfg.Workers)
	for range cfg.Workers {
		s.wg.Add(1)
		go s.work(jobs, ready)
	}
	s.wg.Add(1)
	go s.dispatch(jobs, ready)
	return s, nil
}

func (s *Sender) Push(r Record) {
	b, err := json.Marshal(s.Prepare(r))
	if err == nil {
		rctx, rcancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer rcancel()
		err = s.rdb.RPush(rctx, s.cfg.RedisKey, b).Err()
	}
	if err != nil {
		s.lost.Add(1)
		logger.Error(fmt.Sprintf("LogTransaksi %s tidak tercatat: %s", r.TransactionID, err.Error()))
	}
}

func (s *Sender) Failed() int64 { return s.failed.Load() }

func (s *Sender) Lost() int64 { return s.lost.Load() }

func (s *Sender) Close(ctx context.Context) {
	s.once.Do(func() {
		close(s.stop)
		done := make(chan struct{})
		go func() { s.wg.Wait(); close(done) }()
		select {
		case <-done:
		case <-ctx.Done():
			s.cancel()
			<-done
		}
		s.cancel()
	})
}

func (s *Sender) dispatch(jobs chan<- []byte, ready chan struct{}) {
	defer s.wg.Done()
	defer close(jobs)
	for {
		select {
		case <-ready:
		case <-s.stop:
			return
		}

		b, err := s.pop()
		if err == nil && b != nil && !json.Valid(b) {
			s.lost.Add(1)
			logger.Error("LogTransaksi: catatan rusak di Redis dilewati")
			b = nil
		}
		if b == nil {
			ready <- struct{}{}
			if err != nil {
				logger.Error("LogTransaksi: gagal mengambil antrean Redis: " + err.Error())
				s.wait()
			}
			continue
		}
		jobs <- b
	}
}

func (s *Sender) work(jobs <-chan []byte, ready chan<- struct{}) {
	defer s.wg.Done()
	for {
		select {
		case ready <- struct{}{}:
		case <-s.stop:
			return
		}
		b, ok := <-jobs
		if !ok {
			return
		}
		if err := s.send(b); err != nil {
			s.failed.Add(1)
			s.requeue(b)
			if s.down.CompareAndSwap(false, true) {
				logger.Error("LogTransaksi: service log tidak bisa dihubungi, catatan tetap di Redis: " + err.Error())
			}
			s.wait()
			continue
		}
		if s.down.CompareAndSwap(true, false) {
			logger.Info("LogTransaksi: service log kembali bisa dihubungi")
		}
	}
}

func (s *Sender) pop() ([]byte, error) {
	rctx, rcancel := context.WithTimeout(context.Background(), popTimeout+s.cfg.Timeout)
	defer rcancel()
	res, err := s.rdb.BLPop(rctx, popTimeout, s.cfg.RedisKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return []byte(res[1]), nil
}

func (s *Sender) requeue(b []byte) {
	rctx, rcancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
	defer rcancel()
	if err := s.rdb.LPush(rctx, s.cfg.RedisKey, b).Err(); err != nil {
		s.lost.Add(1)
		logger.Error("LogTransaksi: gagal mengembalikan catatan ke Redis: " + err.Error())
	}
}

func (s *Sender) wait() {
	select {
	case <-s.stop:
	case <-time.After(s.cfg.RetryInterval):
	}
}

func mergeHeaders(headers map[string]string, raw string) (map[string]string, error) {
	merged := map[string]string{}
	for k, v := range headers {
		merged[k] = v
	}
	for i, pair := range strings.Split(raw, ",") {
		if strings.TrimSpace(pair) == "" {
			continue
		}
		k, v, ok := strings.Cut(pair, "=")
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if !ok || k == "" {
			return nil, fmt.Errorf("LOG_TRANSAKSI_HEADERS entri ke-%d harus berformat nama=nilai", i+1)
		}
		merged[k] = v
	}
	return merged, nil
}

func (s *Sender) send(body []byte) error {
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, s.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range s.cfg.Headers {
		req.Header.Set(k, v)
	}
	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	if res.StatusCode >= 300 {
		return fmt.Errorf("service log membalas %d", res.StatusCode)
	}
	return nil
}
