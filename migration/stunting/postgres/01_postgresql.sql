-- +goose Up
-- +goose StatementBegin

-- Postgres tidak membuat schema secara otomatis, jadi ini harus lebih dulu
-- daripada CREATE TABLE mana pun.
CREATE SCHEMA IF NOT EXISTS jakantro;

-- 1. Tabel orangtua
CREATE TABLE IF NOT EXISTS jakantro.orangtua (
    id character(36) NOT NULL,
    satusehat_id character(36) DEFAULT NULL,
    id_posyandu character(36) NOT NULL,
    no_kk character varying(16) NOT NULL,
    nama_ayah character varying(255) NOT NULL,
    nama_ibu character varying(255) NOT NULL,
    nik character varying(16) NOT NULL,
    telepon character varying(255) NOT NULL,
    rt character(3) NOT NULL,
    rw character(3) NOT NULL,
    alamat text NOT NULL,
    kia smallint NOT NULL,
    "createdAt" timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updatedAt" timestamp(6) with time zone,
    "deletedAt" timestamp(6) with time zone,
    source_data character(36) DEFAULT '20f11c5c-5a29-11f0-a136-5749ce66d036'::bpchar,
    usia_hamil smallint,
    kia_bayi_kecil smallint,
    updated_by character varying(100),
    deleted_by character varying(100),
    CONSTRAINT pk_orangtua PRIMARY KEY (id),
    CONSTRAINT uq_orangtua_nik UNIQUE (nik),
    CONSTRAINT uq_orangtua_no_kk UNIQUE (no_kk)
);

-- 2. Tabel anak
CREATE TABLE IF NOT EXISTS jakantro.anak (
    id character(36) NOT NULL,
    satusehat_id character(36) DEFAULT NULL,
    id_orangtua character(36) NOT NULL,
    nama character varying(255) NOT NULL,
    -- nik anak opsional. UNIQUE di Postgres mengizinkan banyak NULL, jadi
    -- banyak anak tanpa nik tetap sah selama nilainya NULL, bukan string kosong.
    nik character varying(16),
    tanggal_lahir date NOT NULL,
    jenis_kelamin character(1) NOT NULL,
    anak_ke smallint NOT NULL,
    imd smallint NOT NULL,
    bb_lahir real NOT NULL,
    tb_lahir real NOT NULL,
    lk_lahir real NOT NULL,
    "createdAt" timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updatedAt" timestamp(6) with time zone,
    "deletedAt" timestamp(6) with time zone,
    source_data character(36) DEFAULT '20f11c5c-5a29-11f0-a136-5749ce66d036'::bpchar NOT NULL,
    status_aktif character varying(20) DEFAULT 'aktif'::character varying,
    updated_by character varying(100),
    deleted_by character varying(100),
    CONSTRAINT pk_anak PRIMARY KEY (id),
    CONSTRAINT uq_anak_nik UNIQUE (nik),
    CONSTRAINT fk_anak_orangtua FOREIGN KEY (id_orangtua) REFERENCES jakantro.orangtua(id) ON DELETE CASCADE
);

-- 3. Tabel kesehatan
CREATE TABLE IF NOT EXISTS jakantro.kesehatan (
    id character(36) NOT NULL,
    id_anak character(36) NOT NULL,
    tanggal_pemantauan date DEFAULT '2025-02-16'::date NOT NULL,
    tbc_batuk smallint,
    tbc_demam smallint,
    tbc_bb smallint,
    tbc_kontak smallint,
    layanan_asi_eks smallint,
    layanan_mpasi smallint,
    layanan_imunisasi smallint,
    layanan_vit_a smallint,
    layanan_obat_cacing smallint,
    layanan_mt_pangan smallint,
    penyuluhan_edukasi smallint,
    penyuluhan_rujukan smallint,
    "createdAt" timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updatedAt" timestamp(6) with time zone,
    "deletedAt" timestamp(6) with time zone,
    source_data character(36) DEFAULT '20f11c5c-5a29-11f0-a136-5749ce66d036'::bpchar,
    CONSTRAINT pk_kesehatan PRIMARY KEY (id),
    CONSTRAINT fk_kesehatan_anak FOREIGN KEY (id_anak) REFERENCES jakantro.anak(id) ON DELETE CASCADE
);

-- 4. Tabel kunjungan (sebelumnya pengukuran)
CREATE TABLE IF NOT EXISTS jakantro.kunjungan (
    id character(36) NOT NULL,
    id_anak character(36) NOT NULL,
    tanggal_pengukuran date DEFAULT '2025-02-16'::date NOT NULL,
    cara_ukur character varying(50),
    berat_badan double precision,
    tinggi_badan double precision,
    lingkar_lengan double precision,
    lingkar_kepala double precision,
    lingkar_dada double precision,
    asi_bulan_1 smallint,
    asi_bulan_2 smallint,
    asi_bulan_3 smallint,
    asi_bulan_4 smallint,
    asi_bulan_5 smallint,
    asi_bulan_6 smallint,
    vit_biru smallint,
    "createdAt" timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updatedAt" timestamp(6) with time zone,
    "deletedAt" timestamp(6) with time zone,
    source_data character(36) DEFAULT '20f11c5c-5a29-11f0-a136-5749ce66d036'::bpchar,
    pitting_edema smallint,
    kelas_ibu_balita smallint,
    asi_bulan_0 smallint,
    vit_merah smallint,
    status_bbu_sigizi character varying(255),
    status_tbu_sigizi character varying(255),
    status_bbtb_sigizi character varying(255),
    zscore_bbu_sigizi numeric(5,2),
    zscore_tbu_sigizi numeric(5,2),
    zscore_bbtb_sigizi numeric(5,2),
    status_bbu_whoantro character varying(255),
    status_tbu_whoantro character varying(255),
    status_bbtb_whoantro character varying(255),
    zscore_bbu_whoantro numeric(5,2),
    zscore_tbu_whoantro numeric(5,2),
    zscore_bbtb_whoantro numeric(5,2),
    updated_by character varying(100),
    deleted_by character varying(100),
    CONSTRAINT pk_kunjungan PRIMARY KEY (id),
    CONSTRAINT fk_kunjungan_anak FOREIGN KEY (id_anak) REFERENCES jakantro.anak(id) ON DELETE CASCADE
);

-- Indeks
--
-- Postgres membuat indeks otomatis untuk PRIMARY KEY dan UNIQUE, tetapi TIDAK
-- untuk kolom foreign key. Tanpa indeks di bawah ini, setiap pencarian anak
-- lewat orangtuanya memindai seluruh tabel, dan setiap DELETE pada baris induk
-- memindai seluruh tabel turunan untuk memeriksa ON DELETE CASCADE.

-- Melayani dua pola sekaligus: pencarian anak lewat (id_orangtua, anak_ke) di
-- CreateAnakTx, dan pemuatan relasi Anak lewat id_orangtua saja di ListAnak
-- (kolom terkiri sudah cukup untuk itu).
CREATE INDEX IF NOT EXISTS idx_anak_orangtua_urutan
    ON jakantro.anak (id_orangtua, anak_ke);

CREATE INDEX IF NOT EXISTS idx_kesehatan_anak
    ON jakantro.kesehatan (id_anak);

CREATE INDEX IF NOT EXISTS idx_kunjungan_anak
    ON jakantro.kunjungan (id_anak);

-- Comments
COMMENT ON COLUMN jakantro.orangtua.id IS 'UUID';
COMMENT ON COLUMN jakantro.anak.id IS 'UUID';
COMMENT ON COLUMN jakantro.anak.id_orangtua IS 'UUID Ref: orangtua, ondelete cascade';
COMMENT ON COLUMN jakantro.kesehatan.id IS 'UUID';
COMMENT ON COLUMN jakantro.kesehatan.id_anak IS 'UUID Ref: anak, ondelete cascade';
COMMENT ON COLUMN jakantro.kunjungan.id IS 'UUID';
COMMENT ON COLUMN jakantro.kunjungan.id_anak IS 'UUID Ref: anak, ondelete cascade';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS jakantro.kunjungan;
DROP TABLE IF EXISTS jakantro.kesehatan;
DROP TABLE IF EXISTS jakantro.anak;
DROP TABLE IF EXISTS jakantro.orangtua;

-- +goose StatementEnd