-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS stunting;

CREATE TABLE IF NOT EXISTS stunting.faskes (
    id varchar(36) NOT NULL,
    id_induk varchar(36) DEFAULT NULL,
    satusehat_id varchar(36) DEFAULT NULL,
    nama varchar(100) NOT NULL,
    jenis varchar(36) NOT NULL,
    wilayah varchar(13),
    alamat text DEFAULT NULL,
    nomor_telpon varchar(36) DEFAULT NULL,
    email varchar(72) DEFAULT NULL,
    status smallint DEFAULT 1,
    created_at    timestamptz  NOT NULL DEFAULT now(),
    updated_at    timestamptz,

    CONSTRAINT pk_faskes PRIMARY KEY(id),
    CONSTRAINT uq_faskes_satusehat_id UNIQUE(satusehat_id),

    CONSTRAINT ck_faskes_status CHECK (status IN (0, 1))
);

CREATE TABLE IF NOT EXISTS stunting.posyandu (
    id varchar(36) NOT NULL,
    id_puskesmas varchar(36),
    nama character varying(255) NOT NULL,
    telepon character varying(255) DEFAULT NULL,
    alamat character varying(255),
    id_kelurahan character varying(36),
    rt character varying(3) NOT NULL,
    rw character varying(3) NOT NULL,
    "createdAt" timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updatedAt" timestamp(6) with time zone,
    "deletedAt" timestamp(6) with time zone,
    nama_pic character varying(100),

    CONSTRAINT pk_posyandu PRIMARY KEY (id),
    CONSTRAINT fk_posyandu_faskes FOREIGN KEY (id_puskesmas) REFERENCES stunting.faskes(id) ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS stunting.orangtua (
    id varchar(36) NOT NULL,
    satusehat_id varchar(36) DEFAULT NULL,
    id_posyandu varchar(36),
    no_kk character varying(16),
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
    CONSTRAINT uq_orangtua_no_kk UNIQUE (no_kk),
    CONSTRAINT uq_orangtua_satusehat_id UNIQUE (satusehat_id),
    CONSTRAINT fk_orangtua_posyandu FOREIGN KEY (id_posyandu) REFERENCES stunting.posyandu(id) ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS stunting.anak (
    id varchar(36) NOT NULL,
    satusehat_id varchar(36) DEFAULT NULL,
    id_orangtua varchar(36),
    nama character varying(255) NOT NULL,
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
    CONSTRAINT uq_anak_satusehat_id UNIQUE (satusehat_id),
    CONSTRAINT fk_anak_orangtua FOREIGN KEY (id_orangtua) REFERENCES stunting.orangtua(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS stunting.kesehatan (
    id varchar(36) NOT NULL,
    id_anak varchar(36) NOT NULL,
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
    CONSTRAINT fk_kesehatan_anak FOREIGN KEY (id_anak) REFERENCES stunting.anak(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS stunting.episode (
    id varchar(36) NOT NULL,
    id_anak varchar(36) NOT NULL,
    id_faskes varchar(36),
    satusehat_id varchar(36),
    system character varying(255),
    kode character varying(50),
    display character varying(255),
    status character varying(30),
    mulai date,
    selesai date,
    "createdAt" timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updatedAt" timestamp(6) with time zone,
    "deletedAt" timestamp(6) with time zone,

    CONSTRAINT pk_episode PRIMARY KEY (id),
    CONSTRAINT uq_episode_satusehat_id UNIQUE (satusehat_id),
    CONSTRAINT fk_episode_anak FOREIGN KEY (id_anak) REFERENCES stunting.anak(id) ON DELETE CASCADE,
    CONSTRAINT fk_episode_faskes FOREIGN KEY (id_faskes) REFERENCES stunting.faskes(id) ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS stunting.kunjungan (
    id varchar(36) NOT NULL,
    satusehat_id varchar(36),
    id_anak varchar(36) NOT NULL,
    tanggal_pengukuran date DEFAULT '2025-02-16'::date NOT NULL,
    tanggal_selesai date,
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
    id_faskes varchar(36),
    id_episode varchar(36),
    ref_episode varchar(36),
    id_rujukan varchar(36),
    ref_rujukan varchar(36),
    stunting smallint,
    CONSTRAINT pk_kunjungan PRIMARY KEY (id),
    CONSTRAINT uq_kunjungan_satusehat_id UNIQUE (satusehat_id),
    CONSTRAINT fk_kunjungan_anak FOREIGN KEY (id_anak) REFERENCES stunting.anak(id) ON DELETE CASCADE,
    CONSTRAINT fk_kunjungan_faskes FOREIGN KEY (id_faskes) REFERENCES stunting.faskes(id) ON DELETE NO ACTION,
    CONSTRAINT fk_kunjungan_episode FOREIGN KEY (id_episode) REFERENCES stunting.episode(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS stunting.diagnosa (
    id varchar(36) NOT NULL,
    id_anak varchar(36) NOT NULL,
    id_kunjungan varchar(36),
    ref_encounter varchar(36),
    satusehat_id varchar(36),
    jenis character varying(20) NOT NULL DEFAULT 'diagnosis',
    system character varying(255) NOT NULL,
    kode character varying(50) NOT NULL,
    display character varying(255),
    kategori character varying(50),
    kritikalitas character varying(30),
    clinical_status character varying(30),
    verification_status character varying(30),
    onset date,
    tanggal_catat date,
    "createdAt" timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updatedAt" timestamp(6) with time zone,
    "deletedAt" timestamp(6) with time zone,

    CONSTRAINT pk_diagnosa PRIMARY KEY (id),
    CONSTRAINT ck_diagnosa_jenis CHECK (jenis IN ('diagnosis', 'alergi')),
    CONSTRAINT uq_diagnosa_satusehat_id UNIQUE (satusehat_id),
    CONSTRAINT fk_diagnosa_anak FOREIGN KEY (id_anak) REFERENCES stunting.anak(id) ON DELETE CASCADE,
    CONSTRAINT fk_diagnosa_kunjungan FOREIGN KEY (id_kunjungan) REFERENCES stunting.kunjungan(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS stunting.observasi (
    id varchar(36) NOT NULL,
    id_anak varchar(36) NOT NULL,
    id_kunjungan varchar(36),
    id_induk varchar(36),
    ref_encounter varchar(36),
    satusehat_id varchar(36),
    system character varying(255) NOT NULL,
    kode character varying(50) NOT NULL,
    display character varying(255),
    kategori character varying(50),
    nilai_angka numeric(12,4),
    satuan character varying(50),
    nilai_teks text,
    nilai_kode character varying(50),
    nilai_kode_system character varying(255),
    nilai_display character varying(255),
    interpretasi character varying(50),
    tanggal timestamp(6) with time zone,
    "createdAt" timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updatedAt" timestamp(6) with time zone,
    "deletedAt" timestamp(6) with time zone,

    CONSTRAINT pk_observasi PRIMARY KEY (id),
    CONSTRAINT uq_observasi_satusehat_id UNIQUE (satusehat_id),
    CONSTRAINT fk_observasi_anak FOREIGN KEY (id_anak) REFERENCES stunting.anak(id) ON DELETE CASCADE,
    CONSTRAINT fk_observasi_kunjungan FOREIGN KEY (id_kunjungan) REFERENCES stunting.kunjungan(id) ON DELETE SET NULL,
    CONSTRAINT fk_observasi_induk FOREIGN KEY (id_induk) REFERENCES stunting.observasi(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS stunting.layanan (
    id varchar(36) NOT NULL,
    id_anak varchar(36) NOT NULL,
    id_kunjungan varchar(36),
    ref_encounter varchar(36),
    satusehat_id varchar(36),
    jenis character varying(30) NOT NULL,
    system character varying(255),
    kode character varying(50),
    display character varying(255),
    kategori character varying(50),
    status character varying(30),
    jumlah numeric(12,4),
    satuan character varying(50),
    tanggal timestamp(6) with time zone,
    catatan text,
    "createdAt" timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updatedAt" timestamp(6) with time zone,
    "deletedAt" timestamp(6) with time zone,

    CONSTRAINT pk_layanan PRIMARY KEY (id),
    CONSTRAINT uq_layanan_satusehat_id UNIQUE (satusehat_id),
    CONSTRAINT ck_layanan_jenis CHECK (jenis IN ('procedure', 'medication_dispense', 'nutrition_order', 'immunization', 'service_request')),
    CONSTRAINT fk_layanan_anak FOREIGN KEY (id_anak) REFERENCES stunting.anak(id) ON DELETE CASCADE,
    CONSTRAINT fk_layanan_kunjungan FOREIGN KEY (id_kunjungan) REFERENCES stunting.kunjungan(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS stunting.rujukan (
    id varchar(36) NOT NULL,
    id_anak varchar(36) NOT NULL,
    id_kunjungan varchar(36),
    ref_encounter varchar(36),
    satusehat_id varchar(36),
    jenis character varying(20) NOT NULL,
    id_faskes_asal varchar(36),
    id_faskes_tujuan varchar(36),
    ref_faskes_asal varchar(36),
    ref_faskes_tujuan varchar(36),
    system character varying(255),
    kode character varying(50),
    display character varying(255),
    status character varying(30),
    prioritas character varying(20),
    alasan text,
    tanggal timestamp(6) with time zone,
    "createdAt" timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updatedAt" timestamp(6) with time zone,
    "deletedAt" timestamp(6) with time zone,

    CONSTRAINT pk_rujukan PRIMARY KEY (id),
    CONSTRAINT uq_rujukan_satusehat_id UNIQUE (satusehat_id),
    CONSTRAINT ck_rujukan_jenis CHECK (jenis IN ('rujukan', 'rujuk_balik', 'internal')),
    CONSTRAINT fk_rujukan_anak FOREIGN KEY (id_anak) REFERENCES stunting.anak(id) ON DELETE CASCADE,
    CONSTRAINT fk_rujukan_kunjungan FOREIGN KEY (id_kunjungan) REFERENCES stunting.kunjungan(id) ON DELETE SET NULL,
    CONSTRAINT fk_rujukan_asal FOREIGN KEY (id_faskes_asal) REFERENCES stunting.faskes(id) ON DELETE NO ACTION,
    CONSTRAINT fk_rujukan_tujuan FOREIGN KEY (id_faskes_tujuan) REFERENCES stunting.faskes(id) ON DELETE NO ACTION
);

CREATE INDEX IF NOT EXISTS idx_orangtua_id_posyandu
    ON stunting.orangtua (id_posyandu) WHERE id_posyandu IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_anak_orangtua_urutan
    ON stunting.anak (id_orangtua, anak_ke);

CREATE INDEX IF NOT EXISTS idx_kesehatan_anak
    ON stunting.kesehatan (id_anak);

CREATE INDEX IF NOT EXISTS idx_kunjungan_anak
    ON stunting.kunjungan (id_anak, "createdAt" DESC);

CREATE INDEX IF NOT EXISTS idx_faskes_wilayah ON stunting.faskes (wilayah);
CREATE INDEX IF NOT EXISTS idx_faskes_nomor_telpon ON stunting.faskes (nomor_telpon) where nomor_telpon IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_faskes_nomor_email ON stunting.faskes (email) where email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_faskes_nama_lower ON stunting.faskes (lower(nama));
CREATE INDEX IF NOT EXISTS idx_faskes_id_induk ON stunting.faskes (id_induk) WHERE id_induk IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_posyandu_id_puskesmas ON stunting.posyandu (id_puskesmas);
CREATE INDEX IF NOT EXISTS idx_posyandu_id_kelurahan ON stunting.posyandu (id_kelurahan);

CREATE INDEX IF NOT EXISTS idx_kunjungan_faskes ON stunting.kunjungan (id_faskes) WHERE id_faskes IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_kunjungan_stunting ON stunting.kunjungan (stunting) WHERE stunting IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_diagnosa_anak ON stunting.diagnosa (id_anak);
CREATE INDEX IF NOT EXISTS idx_diagnosa_kunjungan ON stunting.diagnosa (id_kunjungan);
CREATE INDEX IF NOT EXISTS idx_diagnosa_kode ON stunting.diagnosa (kode);

CREATE INDEX IF NOT EXISTS idx_observasi_anak ON stunting.observasi (id_anak, tanggal);
CREATE INDEX IF NOT EXISTS idx_observasi_kunjungan ON stunting.observasi (id_kunjungan);
CREATE INDEX IF NOT EXISTS idx_observasi_kode ON stunting.observasi (kode);
CREATE INDEX IF NOT EXISTS idx_observasi_induk ON stunting.observasi (id_induk) WHERE id_induk IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_layanan_anak ON stunting.layanan (id_anak, tanggal);
CREATE INDEX IF NOT EXISTS idx_layanan_kunjungan ON stunting.layanan (id_kunjungan);
CREATE INDEX IF NOT EXISTS idx_layanan_jenis ON stunting.layanan (jenis);
CREATE INDEX IF NOT EXISTS idx_rujukan_anak ON stunting.rujukan (id_anak, tanggal);
CREATE INDEX IF NOT EXISTS idx_rujukan_kunjungan ON stunting.rujukan (id_kunjungan);
CREATE INDEX IF NOT EXISTS idx_rujukan_jenis ON stunting.rujukan (jenis);
CREATE INDEX IF NOT EXISTS idx_rujukan_ref_encounter ON stunting.rujukan (ref_encounter) WHERE id_kunjungan IS NULL;
CREATE INDEX IF NOT EXISTS idx_diagnosa_jenis ON stunting.diagnosa (jenis);
CREATE INDEX IF NOT EXISTS idx_episode_anak ON stunting.episode (id_anak);
CREATE INDEX IF NOT EXISTS idx_kunjungan_episode ON stunting.kunjungan (id_episode) WHERE id_episode IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_kunjungan_ref_episode ON stunting.kunjungan (ref_episode) WHERE id_episode IS NULL;

ALTER TABLE stunting.kunjungan
    ADD CONSTRAINT fk_kunjungan_rujukan FOREIGN KEY (id_rujukan)
    REFERENCES stunting.rujukan(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_kunjungan_rujukan ON stunting.kunjungan (id_rujukan) WHERE id_rujukan IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_kunjungan_ref_rujukan ON stunting.kunjungan (ref_rujukan) WHERE id_rujukan IS NULL;

CREATE INDEX IF NOT EXISTS idx_diagnosa_ref_encounter ON stunting.diagnosa (ref_encounter) WHERE id_kunjungan IS NULL;
CREATE INDEX IF NOT EXISTS idx_observasi_ref_encounter ON stunting.observasi (ref_encounter) WHERE id_kunjungan IS NULL;
CREATE INDEX IF NOT EXISTS idx_layanan_ref_encounter ON stunting.layanan (ref_encounter) WHERE id_kunjungan IS NULL;

COMMENT ON COLUMN stunting.orangtua.id IS 'UUID';
COMMENT ON COLUMN stunting.anak.id IS 'UUID';
COMMENT ON COLUMN stunting.anak.id_orangtua IS 'UUID Ref: orangtua, ondelete cascade';
COMMENT ON COLUMN stunting.kesehatan.id IS 'UUID';
COMMENT ON COLUMN stunting.kesehatan.id_anak IS 'UUID Ref: anak, ondelete cascade';
COMMENT ON COLUMN stunting.kunjungan.id IS 'UUID';
COMMENT ON COLUMN stunting.kunjungan.id_anak IS 'UUID Ref: anak, ondelete cascade';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE IF EXISTS stunting.kunjungan DROP CONSTRAINT IF EXISTS fk_kunjungan_rujukan;

DROP TABLE IF EXISTS stunting.rujukan;
DROP TABLE IF EXISTS stunting.layanan;
DROP TABLE IF EXISTS stunting.observasi;
DROP TABLE IF EXISTS stunting.diagnosa;
DROP TABLE IF EXISTS stunting.kunjungan;
DROP TABLE IF EXISTS stunting.episode;
DROP TABLE IF EXISTS stunting.kesehatan;
DROP TABLE IF EXISTS stunting.anak;
DROP TABLE IF EXISTS stunting.orangtua;
DROP TABLE IF EXISTS stunting.posyandu;
DROP TABLE IF EXISTS stunting.faskes;

-- +goose StatementEnd
