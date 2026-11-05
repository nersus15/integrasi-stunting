-- Data awal benchmark. Pola id harus sama dengan tests/benchmark/beban/data.go.

INSERT INTO stunting.faskes (id, satusehat_id, nama, jenis, wilayah, status)
SELECT 'BPKM-' || j, '9' || lpad(j::text, 5, '0'), 'Puskesmas Bench ' || j, 'puskesmas', '3175040', 1
FROM generate_series(1, :n_pkm) j;

INSERT INTO stunting.faskes (id, satusehat_id, nama, jenis, wilayah, status)
SELECT 'BRS-' || r, '8' || lpad(r::text, 5, '0'), 'RS Bench ' || r, 'rs', '3175000', 1
FROM generate_series(1, :n_rs) r;

INSERT INTO stunting.posyandu (id, id_puskesmas, id_kelurahan, nama, rt, rw, alamat)
SELECT 'BPOS-' || j || '-' || k, 'BPKM-' || j, '317504100' || k, 'Posyandu Bench ' || j || '-' || k, '001', '002', 'Jl. Bench'
FROM generate_series(1, :n_pkm) j, generate_series(1, 5) k;

INSERT INTO stunting.orangtua (id, satusehat_id, id_posyandu, no_kk, nama_ayah, nama_ibu, nik, telepon, rt, rw, alamat, kia, source_data, created_by)
SELECT '00000000-0000-4000-a000-' || lpad(i::text, 12, '0'),
	CASE WHEN i % 2 = 0 THEN 'R' || lpad(i::text, 11, '0') END,
	'BPOS-' || (i % :n_pkm + 1) || '-' || (i % 5 + 1),
	'32' || lpad(i::text, 14, '0'), 'Ayah ' || i, 'Ibu ' || i, '31' || lpad(i::text, 14, '0'),
	'0812' || lpad(i::text, 8, '0'), '001', '002', 'Jl. Bench No. ' || i, 1, 'jakantro', 'jakantro'
FROM generate_series(1, :n_anak) i;

INSERT INTO stunting.anak (id, satusehat_id, id_orangtua, nama, nik, tanggal_lahir, jenis_kelamin, anak_ke, imd, bb_lahir, tb_lahir, lk_lahir, source_data, created_by)
SELECT '00000000-0000-4000-b000-' || lpad(i::text, 12, '0'), 'P' || lpad(i::text, 11, '0'),
	'00000000-0000-4000-a000-' || lpad(i::text, 12, '0'), 'Anak ' || i, '33' || lpad(i::text, 14, '0'),
	date '2021-11-01' + (i % 1600), CASE WHEN i % 2 = 0 THEN 'L' ELSE 'P' END, 1, 1, 3.0, 49.0, 34.0, 'jakantro', 'jakantro'
FROM generate_series(1, :n_anak) i;

-- kunjungan ke-1 dari puskesmas (punya satusehat_id), sisanya dari posyandu
INSERT INTO stunting.kunjungan (id, satusehat_id, id_anak, id_faskes, tanggal_pengukuran, cara_ukur,
	berat_badan, tinggi_badan, lingkar_lengan, lingkar_kepala, status_bbu, status_tbu, status_bbtb,
	zscore_bbu, zscore_tbu, zscore_bbtb, stunting, source_data, created_by, "createdAt")
SELECT '00000000-0000-4000-c' || lpad(k::text, 3, '0') || '-' || lpad(i::text, 12, '0'),
	CASE WHEN k = 1 THEN 'E' || lpad(i::text, 11, '0') END,
	'00000000-0000-4000-b000-' || lpad(i::text, 12, '0'),
	CASE WHEN k = 1 THEN 'BPKM-' || (i % :n_pkm + 1) END,
	date '2026-01-01' - (k * 30), 'telentang', 7 + k * 0.3, 70 + k * 0.8, 12.5, 44.0,
	'Berat Badan Normal', 'Pendek', 'Gizi Baik', -1.2, -2.3, -0.4,
	CASE WHEN k = 1 THEN 1 END,
	CASE WHEN k = 1 THEN 'faskes:9' || lpad((i % :n_pkm + 1)::text, 5, '0') ELSE 'jakantro' END,
	CASE WHEN k = 1 THEN '9' || lpad((i % :n_pkm + 1)::text, 5, '0') ELSE 'jakantro' END,
	now() - (k || ' days')::interval
FROM generate_series(1, :n_anak) i, generate_series(1, :kunj_per_anak) k;

INSERT INTO stunting.observasi (id, id_anak, id_kunjungan, satusehat_id, system, kode, display, kategori, nilai_angka, satuan, tanggal)
SELECT '00000000-0000-4000-d00' || o || '-' || lpad(i::text, 12, '0'),
	'00000000-0000-4000-b000-' || lpad(i::text, 12, '0'),
	'00000000-0000-4000-c001-' || lpad(i::text, 12, '0'),
	'O' || o || lpad(i::text, 10, '0'), 'http://loinc.org',
	CASE o WHEN 1 THEN '29463-7' ELSE '8306-3' END, CASE o WHEN 1 THEN 'Body weight' ELSE 'Body height' END,
	'vital-signs', CASE o WHEN 1 THEN 7.3 ELSE 70.8 END, CASE o WHEN 1 THEN 'kg' ELSE 'cm' END, now() - interval '30 days'
FROM generate_series(1, :n_anak) i, generate_series(1, 2) o;

INSERT INTO stunting.rujukan (id, id_anak, id_kunjungan, satusehat_id, jenis, id_faskes_asal, id_faskes_tujuan,
	ref_faskes_asal, ref_faskes_tujuan, status, prioritas, tanggal)
SELECT '00000000-0000-4000-e000-' || lpad(i::text, 12, '0'),
	'00000000-0000-4000-b000-' || lpad(i::text, 12, '0'),
	'00000000-0000-4000-c001-' || lpad(i::text, 12, '0'),
	'S' || lpad(i::text, 11, '0'), 'rujukan',
	'BPKM-' || (i % :n_pkm + 1), 'BRS-' || (i % :n_rs + 1),
	'9' || lpad((i % :n_pkm + 1)::text, 5, '0'), '8' || lpad((i % :n_rs + 1)::text, 5, '0'),
	'active', 'routine', now() - interval '20 days'
FROM generate_series(1, :n_anak) i
WHERE i % 100 < :persen_rujukan;

INSERT INTO access.auth_users (username, password_hash, groups, roles)
VALUES ('bench-jakantro', '-', '{jakantro}', '{read,write}');
INSERT INTO access.auth_apikey (api_key, user_id)
SELECT :'kunci_jakantro', id FROM access.auth_users WHERE username = 'bench-jakantro';

INSERT INTO access.auth_users (username, password_hash, groups, roles)
SELECT 'bench-faskes-' || satusehat_id, '-', ARRAY['orgid:' || satusehat_id], '{read,write}'
FROM stunting.faskes WHERE id LIKE 'BPKM-%' OR id LIKE 'BRS-%';
INSERT INTO access.auth_apikey (api_key, user_id)
SELECT :'kunci_prefix' || substr(username, 14), id FROM access.auth_users WHERE username LIKE 'bench-faskes-%';
