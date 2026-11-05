package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type permintaan struct {
	method, path, kunci string
	body                any
}

type operasi struct {
	nama  string
	bobot int
	buat  func(p *populasi) permintaan
}

func daftarOperasi(bobot func(string) int) []operasi {
	all := []operasi{
		{"get_anak", bobot("BOBOT_GET_ANAK"), func(p *populasi) permintaan {
			return permintaan{"GET", "/anak/" + idAnak(p.acak()), p.kunciJakantro, nil}
		}},
		{"get_kunjungan_anak", bobot("BOBOT_GET_KUNJUNGAN_ANAK"), func(p *populasi) permintaan {
			return permintaan{"GET", "/anak/" + idAnak(p.acak()) + "/kunjungan", p.kunciJakantro, nil}
		}},
		{"get_summary", bobot("BOBOT_GET_SUMMARY"), func(p *populasi) permintaan {
			return permintaan{"GET", "/anak/" + idAnak(p.acak()) + "/summary", p.kunciJakantro, nil}
		}},
		{"get_orangtua", bobot("BOBOT_GET_ORANGTUA"), func(p *populasi) permintaan {
			return permintaan{"GET", "/orangtua/" + idOrangtua(p.acak()), p.kunciJakantro, nil}
		}},
		// pemegang akses membaca anaknya: puskesmas posyandu, atau RS tujuan rujukan
		{"get_faskes", bobot("BOBOT_GET_FASKES"), func(p *populasi) permintaan {
			i := p.acak()
			return permintaan{"GET", "/anak/" + idAnak(i), p.kunciFaskes(p.orgPemegang(i)), nil}
		}},
		{"get_faskes_orangtua", bobot("BOBOT_GET_FASKES_ORANGTUA"), func(p *populasi) permintaan {
			i := p.acak()
			return permintaan{"GET", "/orangtua/" + idOrangtua(i), p.kunciFaskes(p.orgPemegang(i)), nil}
		}},
		{"post_registrasi", bobot("BOBOT_REGISTRASI"), func(p *populasi) permintaan {
			return permintaan{"POST", "/kunjungan", p.kunciJakantro, p.registrasi()}
		}},
		{"post_kunjungan", bobot("BOBOT_KUNJUNGAN"), func(p *populasi) permintaan {
			return permintaan{"POST", "/kunjungan", p.kunciJakantro, p.kunjunganSaja(p.acak())}
		}},
		{"post_kesehatan", bobot("BOBOT_KESEHATAN"), func(p *populasi) permintaan {
			return permintaan{"POST", "/kesehatan", p.kunciJakantro, p.kesehatan(p.acak())}
		}},
		{"put_anak", bobot("BOBOT_PUT_ANAK"), func(p *populasi) permintaan {
			return permintaan{"PUT", "/anak", p.kunciJakantro, putAnak(p.acak())}
		}},
		{"put_orangtua", bobot("BOBOT_PUT_ORANGTUA"), func(p *populasi) permintaan {
			return permintaan{"PUT", "/orangtua", p.kunciJakantro, putOrangtua(p.acak())}
		}},
		{"faskes_pemeriksaan", bobot("BOBOT_FASKES_PEMERIKSAAN"), func(p *populasi) permintaan {
			i := p.acak()
			return permintaan{"POST", "/faskes/pemeriksaan", p.kunciFaskes(p.orgPemegang(i)), p.pemeriksaan(i)}
		}},
		{"faskes_put_kunjungan", bobot("BOBOT_FASKES_PUT_KUNJUNGAN"), func(p *populasi) permintaan {
			i := p.acak()
			return permintaan{"PUT", "/faskes/kunjungan", p.kunciFaskes(p.orgPkm(i)), putKunjunganFaskes(i)}
		}},
	}
	var dipakai []operasi
	for _, o := range all {
		if o.bobot > 0 {
			dipakai = append(dipakai, o)
		}
	}
	return dipakai
}

type klienHTTP struct {
	base string
	c    *http.Client
}

func klienBaru(base string, timeout time.Duration, inflight int) *klienHTTP {
	t := &http.Transport{MaxIdleConns: inflight, MaxIdleConnsPerHost: inflight, IdleConnTimeout: 90 * time.Second}
	return &klienHTTP{base: base, c: &http.Client{Timeout: timeout, Transport: t}}
}

func (k *klienHTTP) kirim(r permintaan) (int, string) {
	var body io.Reader
	if r.body != nil {
		b, _ := json.Marshal(r.body)
		body = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(r.method, k.base+r.path, body)
	req.Header.Set("X-API-Key", r.kunci)
	req.Header.Set("Content-Type", "application/json")
	res, err := k.c.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer res.Body.Close()
	content, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(content)
}

// setiap operasi harus berhasil sekali sebelum diberi beban
func cekOperasi(k *klienHTTP, p *populasi, ops []operasi) error {
	for _, o := range ops {
		code, content := k.kirim(o.buat(p))
		if code < 200 || code >= 300 {
			return fmt.Errorf("operasi %s gagal sebelum beban: %d %s", o.nama, code, potong(content, 400))
		}
	}
	return nil
}

type pemilih struct {
	ops   []operasi
	batas []int
	total int
}

func pemilihBaru(ops []operasi) *pemilih {
	p := &pemilih{ops: ops}
	for _, o := range ops {
		p.total += o.bobot
		p.batas = append(p.batas, p.total)
	}
	return p
}

func (p *pemilih) ambil() operasi {
	n := rand.IntN(p.total)
	return p.ops[sort.SearchInts(p.batas, n+1)]
}

type hasilTahap struct {
	Laju     float64
	Tercapai float64
	Durasi   time.Duration
	Total    ringkas
	PerOp    map[string]ringkas
	Contoh   map[string]string
	Sumber   sumberDaya
	Kafka    *hasilKafka `json:",omitempty"`
	Jenuh    string
}

// open-loop: request dijadwalkan pada laju tetap, tidak menunggu respons sebelumnya
func bebanHTTP(ctx context.Context, k *klienHTTP, p *populasi, pil *pemilih, laju float64, durasi time.Duration, maksInflight int, c *catatan) float64 {
	var inflight atomic.Int64
	var wg sync.WaitGroup
	jeda := time.Duration(float64(time.Second) / laju)
	start := time.Now()
	done := start.Add(durasi)
	var selesaiTepat atomic.Int64
	for n := 0; ; n++ {
		jadwal := start.Add(time.Duration(n) * jeda)
		if !jadwal.Before(done) || ctx.Err() != nil {
			break
		}
		if d := time.Until(jadwal); d > 0 {
			time.Sleep(d)
		}
		op := pil.ambil()
		if inflight.Load() >= int64(maksInflight) {
			c.tambah(op.nama, 0, -1, "")
			continue
		}
		inflight.Add(1)
		wg.Add(1)
		go func(op operasi, jadwal time.Time) {
			defer wg.Done()
			defer inflight.Add(-1)
			code, content := k.kirim(op.buat(p))
			c.tambah(op.nama, time.Since(jadwal), code, content)
			if time.Now().Before(done) && code >= 200 && code < 300 {
				selesaiTepat.Add(1)
			}
		}(op, jadwal)
	}
	wg.Wait()
	return float64(selesaiTepat.Load()) / durasi.Seconds()
}
