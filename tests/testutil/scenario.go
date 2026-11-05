package testutil

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
)

const (
	Scenario1IhsAnak     = "P99001100022"
	Scenario1NikAnak     = "3175042003240007"
	Scenario1NikIbu      = "3175040202900008"
	Scenario1IdAnak      = "a1b2c3d4-2222-4aaa-8bbb-000000000002"
	Scenario1IdOrtu      = "a1b2c3d4-1111-4aaa-8bbb-000000000001"
	Scenario1IdKunjungan = "a1b2c3d4-3333-4aaa-8bbb-000000000003"
)

func Scenario1File(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(RepoRoot(), "simulasi", "skenario1", name))
}

func Scenario1Seed(db *sql.DB) error {
	seed, err := Scenario1File("00-seed.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(seed))
	return err
}

func Scenario1Cleanup(db *sql.DB, groupKafka string) error {
	perintah := []struct {
		q    string
		args []any
	}{
		{`DELETE FROM stunting.anak WHERE nik = $1 OR satusehat_id = $2 OR id = $3`, []any{Scenario1NikAnak, Scenario1IhsAnak, Scenario1IdAnak}},
		{`DELETE FROM stunting.orangtua WHERE nik = $1`, []any{Scenario1NikIbu}},
		{`DELETE FROM stunting.episode WHERE satusehat_id LIKE 'ep-%sim1%'`, nil},
		{`DELETE FROM kafka.transactions WHERE group_id = $1`, []any{groupKafka}},
	}
	for _, p := range perintah {
		if _, err := db.Exec(p.q, p.args...); err != nil {
			return err
		}
	}
	return nil
}

func (r *Wiring) Scenario1Jakantro() error {
	content, err := Scenario1File("01-jakantro-kunjungan.json")
	if err != nil {
		return err
	}
	_, err = r.Service.CreateKunjungan(content)
	return err
}

func (r *Wiring) Scenario1Send(file ...string) error {
	for _, b := range file {
		content, err := Scenario1File(b)
		if err != nil {
			return err
		}
		r.Kafka.Run(context.Background(), content)
	}
	return nil
}

var Scenario1KafkaFiles = []string{
	"02-kafka-patient.json",
	"03-kafka-relatedperson.json",
	"04-kafka-puskesmas.json",
	"05-kafka-rs.json",
	"06-kafka-rujuk-balik.json",
	"07-kafka-puskesmas-lanjutan.json",
}
