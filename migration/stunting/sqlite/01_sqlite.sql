-- +goose Up

CREATE TABLE IF NOT EXISTS patient (
    id TEXT PRIMARY KEY NOT NULL,
    patient_id TEXT,
    nik TEXT,
    nama TEXT,
    nik_ayah TEXT,
    nik_ibu TEXT,
    nama_ayah TEXT,
    nama_ibu TEXT
);

CREATE INDEX IF NOT EXISTS idx_patient_patient_id ON patient (patient_id);
-- Satu NIK hanya mewakili satu balita, jadi indeksnya harus UNIQUE.
CREATE UNIQUE INDEX IF NOT EXISTS idx_patient_nik ON patient(nik);
CREATE INDEX IF NOT EXISTS idx_patient_nik_ibu ON patient(nik_ibu);
CREATE INDEX IF NOT EXISTS idx_patient_nik_ayah ON patient(nik_ayah);

CREATE TABLE IF NOT EXISTS kunjungan (
    id TEXT NOT NULL,
    tanggal_kunjungan TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    id_balita TEXT NOT NULL,
    berat_badan NUMBER NOT NULL,
    tinggi_badan NUMBER NOT NULL,

    FOREIGN KEY (id_balita) REFERENCES patient(id)
);



CREATE INDEX IF NOT EXISTS idx_kunjungan_id_balita ON kunjungan(id_balita);
CREATE INDEX IF NOT EXISTS idx_kunjungan_tanggal_kunjungan ON kunjungan(tanggal_kunjungan DESC);