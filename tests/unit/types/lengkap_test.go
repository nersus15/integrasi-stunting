package types_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func str(s string) *string { return &s }

func sameName(t *testing.T, label string, a, b reflect.Value) {
	t.Helper()
	for i := 0; i < a.NumField(); i++ {
		fa := a.Type().Field(i)
		if !fa.IsExported() || fa.Anonymous {
			continue
		}
		for j := 0; j < b.NumField(); j++ {
			fb := b.Type().Field(j)
			if !strings.EqualFold(fa.Name, fb.Name) {
				continue
			}
			if fa.Type == fb.Type {
				assert.Equal(t, a.Field(i).Interface(), b.Field(j).Interface(), "%s.%s", label, fa.Name)
			}
		}
	}
}

func TestKonversiNilTetapNil(t *testing.T) {
	var (
		kp  *types.KesehatanPayload
		kjp *types.KunjunganPayload
		ap  *types.AnakPayload
		op  *types.OrangtuaPayload
		fp  *types.FaskesPayload
		pp  *types.PosyanduPayload
		a   *types.Anak
		o   *types.Orangtua
		kj  *types.Kunjungan
		ks  *types.Kesehatan
		f   *types.Faskes
		fd  *types.FaskesDetail
		p   *types.Posyandu
		lp  *types.ListPosyandu
		pd  *types.PosyanduDetail
		la  *types.ListAnak
		ka  *types.KunjunganAnakArray
		kh  *types.KesehatanAnakArray
		sa  *types.SummaryAnak
		ob  *types.Observasi
		dg  *types.Diagnosa
		ly  *types.Layanan
		rj  *types.Rujukan
		rd  *types.RujukanDetail
		ep  *types.Episode
	)

	assert.Nil(t, kp.ToEntity())
	assert.Nil(t, kjp.ToEntity())
	assert.Nil(t, ap.ToEntity())
	assert.Nil(t, op.ToEntity())
	assert.Nil(t, fp.ToEntity())
	assert.Nil(t, pp.ToEntity())
	assert.Nil(t, a.ToPayload())
	assert.Nil(t, o.ToPayload())
	assert.Nil(t, kj.ToPayload())
	assert.Nil(t, ks.ToPayload())
	assert.Nil(t, ob.ToEntity())
	assert.Nil(t, dg.ToEntity())
	assert.Nil(t, ly.ToEntity())
	assert.Nil(t, rj.ToEntity())
	assert.Nil(t, ep.ToEntity())

	assert.Nil(t, a.FromEntity(nil))
	assert.Nil(t, o.FromEntity(nil))
	assert.Nil(t, kj.FromEntity(nil))
	assert.Nil(t, ks.FromEntity(nil))
	assert.Nil(t, f.FromEntity(nil))
	assert.Nil(t, fd.FromEntity(nil))
	assert.Nil(t, p.FromEntity(nil))
	assert.Nil(t, lp.FromEntity(nil))
	assert.Nil(t, pd.FromEntity(nil))
	assert.Nil(t, la.FromEntity(nil))
	assert.Nil(t, ka.FromEntity(nil))
	assert.Nil(t, kh.FromEntity(nil))
	assert.Nil(t, sa.FromEntity(nil))
	assert.Nil(t, ob.FromEntity(nil))
	assert.Nil(t, dg.FromEntity(nil))
	assert.Nil(t, ly.FromEntity(nil))
	assert.Nil(t, rj.FromEntity(nil))
	assert.Nil(t, rd.FromEntity(nil))
	assert.Nil(t, ep.FromEntity(nil))
}

func TestFaskesDanPosyanduFromEntity(t *testing.T) {
	n := 0
	induk := new(entity.Faskes)
	fillEntity(reflect.ValueOf(induk).Elem(), &n)
	fk := new(entity.Faskes)
	fillEntity(reflect.ValueOf(fk).Elem(), &n)
	fk.Induk = induk
	pos := new(entity.Posyandu)
	fillEntity(reflect.ValueOf(pos).Elem(), &n)
	pos.Puskesmas = fk
	fk.Posyandu = []*entity.Posyandu{pos, nil}

	var f *types.Faskes
	sameName(t, "Faskes", reflect.ValueOf(*f.FromEntity(fk)), reflect.ValueOf(*fk))

	var fd *types.FaskesDetail
	detail := fd.FromEntity(fk)
	sameName(t, "FaskesDetail", reflect.ValueOf(*detail), reflect.ValueOf(*fk))
	assert.Equal(t, induk.ID, detail.Induk.Id)

	var lp *types.ListPosyandu
	list := lp.FromEntity(fk)
	assert.Equal(t, fk.ID, list.Faskes.Id)
	require.Len(t, list.Posyandu, 1)
	sameName(t, "Posyandu", reflect.ValueOf(list.Posyandu[0]), reflect.ValueOf(*pos))

	var pd *types.PosyanduDetail
	pdetail := pd.FromEntity(pos)
	assert.Equal(t, pos.ID, pdetail.Posyandu.Id)
	assert.Equal(t, fk.ID, pdetail.Puskesmas.Id)

	b, err := json.Marshal(lp.FromEntity(&entity.Faskes{ID: "x"}))
	require.NoError(t, err)
	assert.Contains(t, string(b), `"posyandu":[]`)
}

func TestFaskesDanPosyanduPayloadToEntity(t *testing.T) {
	n := 0
	fp := new(types.FaskesPayload)
	fillEntity(reflect.ValueOf(fp).Elem(), &n)
	fe := fp.ToEntity()
	sameName(t, "FaskesPayload", reflect.ValueOf(*fp), reflect.ValueOf(*fe))
	assert.Equal(t, *fp.CreatedAt, fe.CreatedAt)

	pp := new(types.PosyanduPayload)
	fillEntity(reflect.ValueOf(pp).Elem(), &n)
	pe := pp.ToEntity()
	sameName(t, "PosyanduPayload", reflect.ValueOf(*pp), reflect.ValueOf(*pe))
	assert.Equal(t, *pp.CreatedAt, pe.CreatedAt)

	fp.CreatedAt, pp.CreatedAt = nil, nil
	assert.True(t, fp.ToEntity().CreatedAt.IsZero(), "biar database yang mengisi")
	assert.True(t, pp.ToEntity().CreatedAt.IsZero())
}

func TestDaftarAnakKunjunganKesehatanMelewatiNil(t *testing.T) {
	anak := &entity.Anak{ID: "a1", Nama: "A",
		Kunjungan: []*entity.Kunjungan{nil, {ID: "k1", IDAnak: "a1"}},
		Kesehatan: []*entity.Kesehatan{{ID: "s1", IDAnak: "a1"}, nil},
	}
	ortu := &entity.Orangtua{ID: "o1", Anak: []*entity.Anak{nil, anak}}

	var la *types.ListAnak
	l := la.FromEntity(ortu)
	assert.Equal(t, "o1", l.Orangtua.Id)
	require.Len(t, l.Anak, 1)
	assert.Equal(t, "a1", l.Anak[0].Id)

	var ka *types.KunjunganAnakArray
	k := ka.FromEntity(anak)
	require.Len(t, k.Kunjungan, 1)
	assert.Equal(t, "k1", k.Kunjungan[0].Id)

	var kh *types.KesehatanAnakArray
	h := kh.FromEntity(anak)
	require.Len(t, h.Kesehatan, 1)
	assert.Equal(t, "s1", h.Kesehatan[0].Id)
}

func TestSummaryAnakMerangkaiRiwayat(t *testing.T) {
	tgl := time.Date(2026, 2, 5, 0, 0, 0, 0, time.UTC)
	faskes := &entity.Faskes{ID: "f1", Nama: "Puskesmas"}
	rujukan := &entity.Rujukan{ID: "r1", IDAnak: "a1", IDKunjungan: str("k0"), Jenis: entity.RujukanKeluar, Tanggal: &tgl}
	gantung := &entity.Rujukan{ID: "r2", IDAnak: "a1"}
	episode := &entity.Episode{ID: "e1", IDAnak: "a1"}

	k1 := &entity.Kunjungan{ID: "k1", IDAnak: "a1", TanggalPengukuran: tgl, IDFaskes: str("f1"), Faskes: faskes,
		IDRujukan: str("r1"), AtasRujukan: rujukan, IDEpisode: str("e1"), Episode: episode,
		Observasi: []*entity.Observasi{{ID: "o1", IDKunjungan: str("k1")}, nil},
		Diagnosa:  []*entity.Diagnosa{{ID: "d1", IDKunjungan: str("k1")}},
		Layanan:   []*entity.Layanan{{ID: "l1", IDKunjungan: str("k1")}},
		Rujukan:   []*entity.Rujukan{rujukan},
	}
	k2 := &entity.Kunjungan{ID: "k2", IDAnak: "a1", TanggalPengukuran: tgl}

	anak := &entity.Anak{ID: "a1",
		Kunjungan: []*entity.Kunjungan{nil, k1, k2},
		Episode:   []*entity.Episode{episode, nil},
		Kesehatan: []*entity.Kesehatan{{ID: "s1"}},
		Observasi: []*entity.Observasi{{ID: "o1", IDKunjungan: str("k1")}, {ID: "o2"}},
		Diagnosa:  []*entity.Diagnosa{{ID: "d2"}, nil},
		Layanan:   []*entity.Layanan{{ID: "l1", IDKunjungan: str("k1")}, {ID: "l2"}},
		Rujukan:   []*entity.Rujukan{rujukan, gantung, nil},
	}

	var s *types.SummaryAnak
	r := s.FromEntity(anak)

	require.Len(t, r.Kunjungan, 2)
	assert.Equal(t, "f1", r.Kunjungan[0].Faskes.Id)
	assert.Equal(t, "e1", r.Kunjungan[0].Episode.Id)
	assert.Equal(t, "r1", r.Kunjungan[0].AtasRujukan.Id)
	assert.Len(t, r.Kunjungan[0].Observasi, 1)
	assert.Nil(t, r.Kunjungan[1].Faskes)

	require.Len(t, r.Rujukan, 2)
	require.Len(t, r.Rujukan[0].Tindakan, 1, "r1 ditindaklanjuti k1")
	assert.Equal(t, "k1", r.Rujukan[0].Tindakan[0].Id)
	assert.Equal(t, "2026-02-05", r.Rujukan[0].Tindakan[0].TanggalPengukuran)
	assert.Equal(t, "f1", r.Rujukan[0].Tindakan[0].Faskes.Id)
	assert.NotNil(t, r.Rujukan[1].Tindakan)
	assert.Empty(t, r.Rujukan[1].Tindakan)

	assert.Len(t, r.Episode, 1)
	assert.Len(t, r.Kesehatan, 1)
	assert.Len(t, r.Layanan, 2)
	assert.Equal(t, []string{"o2"}, idObservasi(r.Menggantung.Observasi))
	assert.Len(t, r.Menggantung.Diagnosa, 1)
	assert.Len(t, r.Menggantung.Layanan, 1)
	assert.Len(t, r.Menggantung.Rujukan, 1)
	assert.Equal(t, "r2", r.Menggantung.Rujukan[0].Id)
}

func idObservasi(list []types.Observasi) []string {
	out := []string{}
	for _, o := range list {
		out = append(out, o.Id)
	}
	return out
}

func TestRujukanDetailMembawaFaskes(t *testing.T) {
	n := 0
	e := new(entity.Rujukan)
	fillEntity(reflect.ValueOf(e).Elem(), &n)
	e.FaskesAsal = &entity.Faskes{ID: "asal"}
	e.FaskesTujuan = &entity.Faskes{ID: "tujuan"}

	var rd *types.RujukanDetail
	d := rd.FromEntity(e)
	sameName(t, "RujukanDetail", reflect.ValueOf(d.Rujukan), reflect.ValueOf(*e))
	assert.Equal(t, "asal", d.FaskesAsal.Id)
	assert.Equal(t, "tujuan", d.FaskesTujuan.Id)
	assert.Equal(t, e.Tanggal.Format(time.RFC3339), *d.Tanggal)
}

func TestWaktuDariStringMenerimaTigaBentuk(t *testing.T) {
	cases := map[string]*time.Time{
		"2026-01-02T03:04:05+07:00": ptrWaktu(time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("", 7*3600))),
		"2026-01-02T03:04:05":       ptrWaktu(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)),
		"2026-01-02":                ptrWaktu(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)),
		"bukan tanggal":             nil,
		"":                          nil,
	}
	for teks, want := range cases {
		o := &types.Observasi{Tanggal: str(teks)}
		result := o.ToEntity().Tanggal
		if want == nil {
			assert.Nil(t, result, teks)
			continue
		}
		if assert.NotNil(t, result, teks) {
			assert.True(t, want.Equal(*result), "%s: %v", teks, result)
		}
	}
	assert.Nil(t, (&types.Observasi{}).ToEntity().Tanggal)
}

func ptrWaktu(t time.Time) *time.Time { return &t }

func TestMapperLabelDanKey(t *testing.T) {
	assert.Equal(t, "Tidur", *types.Pengukuran.ToLabel(1))
	assert.Nil(t, types.Pengukuran.ToLabel(9))
	assert.Equal(t, "2", *types.Pengukuran.ToKey("Berdiri", nil))
	assert.Nil(t, types.Pengukuran.ToKey("Duduk", nil))
	assert.Equal(t, "cadangan", *types.Pengukuran.ToKey("Duduk", str("cadangan")))

	m := types.Mapper{Map: map[any]any{1: nil}, Default: "lainnya"}
	assert.Equal(t, "lainnya", *m.ToLabel(1))
	assert.Equal(t, "lainnya", *m.ToLabel(2))
}

func TestKafkaTransactionBolakBalik(t *testing.T) {
	var kt *types.KafkaTransaction
	var kp *types.KafkaTransactionPayload
	assert.Nil(t, kt.FromEntity(nil))
	assert.Nil(t, kt.ToPayload())
	assert.Nil(t, kp.ToEntity())

	dibuat := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	done := dibuat.Add(time.Hour)
	galat := json.RawMessage(`"gagal"`)
	e := &entity.FailedTransactions{Id: "i", TransactionId: "t", GroupId: "g", PatientId: str("P1"),
		Message: "{}", Error: &galat, Attempt: 3, ResolvedAt: &done, CreatedAt: dibuat}

	dto := kt.FromEntity(e)
	assert.Equal(t, "2026-09-01T10:00:00Z", dto.Created)
	assert.Equal(t, "2026-09-01T11:00:00Z", *dto.Resolved)

	kembali := dto.ToPayload().ToEntity()
	assert.Equal(t, e.CreatedAt, kembali.CreatedAt)
	assert.True(t, e.ResolvedAt.Equal(*kembali.ResolvedAt))
	kembali.ResolvedAt, e.ResolvedAt = nil, nil
	kembali.CreatedAt, e.CreatedAt = time.Time{}, time.Time{}
	assert.Equal(t, e, kembali)

	rusak := &types.KafkaTransaction{Created: "kemarin", Resolved: str("besok")}
	p := rusak.ToPayload()
	assert.WithinDuration(t, time.Now(), p.Created, time.Second)
	assert.Nil(t, p.Resolved)

	e.ResolvedAt = nil
	assert.Nil(t, kt.FromEntity(e).Resolved)
}
