package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type kafkaBench struct {
	broker, topic, group string
	partisi              int
	prod                 *kafka.Producer
	monitor              *kafka.Consumer
}

func siapkanTopic(broker, topic string, partisi int) error {
	adm, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		return err
	}
	defer adm.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, _ = adm.DeleteTopics(ctx, []string{topic})
	time.Sleep(2 * time.Second)
	for range 10 {
		res, err := adm.CreateTopics(ctx, []kafka.TopicSpecification{{Topic: topic, NumPartitions: partisi, ReplicationFactor: 1}})
		if err == nil && (res[0].Error.Code() == kafka.ErrNoError || res[0].Error.Code() == kafka.ErrTopicAlreadyExists) {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("gagal membuat topic %s", topic)
}

func hapusTopic(broker, topic string) error {
	adm, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		return err
	}
	defer adm.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err = adm.DeleteTopics(ctx, []string{topic})
	return err
}

func kafkaBaru(broker, topic, group string, partisi int) (*kafkaBench, error) {
	prod, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker, "linger.ms": 5})
	if err != nil {
		return nil, err
	}
	// group yang sama dengan app tanpa subscribe: hanya membaca offset yang di-commit app
	mon, err := kafka.NewConsumer(&kafka.ConfigMap{"bootstrap.servers": broker, "group.id": group, "enable.auto.commit": false})
	if err != nil {
		prod.Close()
		return nil, err
	}
	go func() {
		for range prod.Events() {
		}
	}()
	return &kafkaBench{broker: broker, topic: topic, group: group, partisi: partisi, prod: prod, monitor: mon}, nil
}

func (k *kafkaBench) tutup() {
	k.prod.Flush(10000)
	k.prod.Close()
	k.monitor.Close()
}

// lag = pesan diproduksi yang belum di-commit app; selesai = total offset yang sudah di-commit
func (k *kafkaBench) posisi() (lag, done int64, err error) {
	var tp []kafka.TopicPartition
	for p := range k.partisi {
		tp = append(tp, kafka.TopicPartition{Topic: &k.topic, Partition: int32(p)})
	}
	com, err := k.monitor.Committed(tp, 5000)
	if err != nil {
		return 0, 0, err
	}
	for _, c := range com {
		_, tinggi, err := k.monitor.QueryWatermarkOffsets(k.topic, c.Partition, 5000)
		if err != nil {
			return 0, 0, err
		}
		o := int64(c.Offset)
		if o < 0 {
			o = 0
		}
		done += o
		lag += tinggi - o
	}
	return lag, done, nil
}

func (k *kafkaBench) kirim(ihs string, message []string) error {
	for _, m := range message {
		for {
			err := k.prod.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &k.topic, Partition: kafka.PartitionAny},
				Key:            []byte(ihs), Value: []byte(m),
			}, nil)
			if err == nil {
				break
			}
			if e, ok := err.(kafka.Error); ok && e.Code() == kafka.ErrQueueFull {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			return err
		}
	}
	return nil
}

type hasilKafka struct {
	LajuTarget     float64
	Diproduksi     int64
	Diproses       int64
	DiprosesPerDtk float64
	LagAwal        int64
	LagAkhir       int64
	PesanGagal     int64
}

// pesan diproduksi dalam satuan perjalanan (6 pesan) supaya urutan per anak sama dengan dunia nyata
func bebanKafka(ctx context.Context, k *kafkaBench, pj *perjalanan, p *populasi, laju float64, durasi time.Duration, db *sql.DB) (*hasilKafka, error) {
	h := &hasilKafka{LajuTarget: laju}
	lagAwal, selesaiAwal, err := k.posisi()
	if err != nil {
		return nil, err
	}
	gagalAwal := hitungGagal(db, k.group)
	h.LagAwal = lagAwal

	n := int64(len(pj.pesan))
	jeda := time.Duration(float64(time.Second) * float64(n) / laju)
	start := time.Now()
	for j := 0; ; j++ {
		jadwal := start.Add(time.Duration(j) * jeda)
		if !jadwal.Before(start.Add(durasi)) || ctx.Err() != nil {
			break
		}
		if d := time.Until(jadwal); d > 0 {
			time.Sleep(d)
		}
		urut := p.baru()
		i := p.acak()
		ihs, message := pj.untuk(urut, p.orgPkm(i), p.orgRs(i))
		if err := k.kirim(ihs, message); err != nil {
			return nil, err
		}
		h.Diproduksi += n
	}
	k.prod.Flush(10000)
	// auto-commit app berjalan tiap 5 detik
	time.Sleep(6 * time.Second)
	lagAkhir, selesaiAkhir, err := k.posisi()
	if err != nil {
		return nil, err
	}
	h.LagAkhir = lagAkhir
	h.Diproses = selesaiAkhir - selesaiAwal
	h.DiprosesPerDtk = float64(h.Diproses) / (time.Since(start).Seconds())
	h.PesanGagal = hitungGagal(db, k.group) - gagalAwal
	return h, nil
}

// tunggu antrean habis supaya tahap berikutnya mulai dari lag nol
func (k *kafkaBench) tungguKosong(limit time.Duration) (time.Duration, int64) {
	start := time.Now()
	for time.Since(start) < limit {
		lag, _, err := k.posisi()
		if err == nil && lag == 0 {
			return time.Since(start), 0
		}
		time.Sleep(2 * time.Second)
	}
	lag, _, _ := k.posisi()
	return time.Since(start), lag
}

func hitungGagal(db *sql.DB, group string) int64 {
	if db == nil {
		return 0
	}
	var n int64
	_ = db.QueryRow(`SELECT count(*) FROM kafka.transactions WHERE group_id = $1`, group).Scan(&n)
	return n
}

type kapasitasKafka struct {
	Pesan         int64
	PesanPerDetik float64
	AnakPerDetik  float64
	Durasi        time.Duration
	Gagal         int64
}

// antrean penuh sekaligus lalu ukur laju commit consumer; inilah batas atas jalur Kafka satu instance
func ukurKapasitasKafka(k *kafkaBench, pj *perjalanan, p *populasi, perjalananN int, db *sql.DB) (*kapasitasKafka, error) {
	gagalAwal := hitungGagal(db, k.group)
	for range perjalananN {
		i := p.acak()
		ihs, message := pj.untuk(p.baru(), p.orgPkm(i), p.orgRs(i))
		if err := k.kirim(ihs, message); err != nil {
			return nil, err
		}
	}
	k.prod.Flush(30000)
	type titik struct {
		t       time.Time
		selesai int64
	}
	var ts []titik
	start := time.Now()
	for time.Since(start) < 15*time.Minute {
		lag, done, err := k.posisi()
		if err != nil {
			return nil, err
		}
		if len(ts) == 0 || done != ts[len(ts)-1].selesai {
			ts = append(ts, titik{time.Now(), done})
		}
		if lag == 0 {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	h := &kapasitasKafka{Pesan: int64(perjalananN * len(pj.pesan)), Durasi: time.Since(start), Gagal: hitungGagal(db, k.group) - gagalAwal}
	if len(ts) >= 3 {
		// titik pertama dan terakhir bisa terpotong interval auto-commit; pakai yang di tengah
		a, z := ts[1], ts[len(ts)-1]
		h.PesanPerDetik = float64(z.selesai-a.selesai) / z.t.Sub(a.t).Seconds()
		h.AnakPerDetik = h.PesanPerDetik / float64(len(pj.pesan))
	}
	return h, nil
}
