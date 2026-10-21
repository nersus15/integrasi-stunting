package service

import (
	"strings"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
)

const (
	pembuatJakantro  = "jakantro"
	pembuatSatusehat = "satusehat"
)

// pembuat masuk created_by, sumber masuk source_data
type asal struct {
	pembuat string
	sumber  string
}

func asalJakantro(label *string) asal {
	if !utils.IsFilled(label) {
		return asal{pembuatJakantro, pembuatJakantro}
	}
	return asal{pembuatJakantro, pembuatJakantro + ":" + strings.TrimSpace(*label)}
}

func asalFaskes(orgid string) asal { return asal{orgid, "faskes:" + orgid} }

// stream: identitas dibuat "satusehat", kunjungan dibuat faskes pengirim Encounter
func asalStream(orgid string) asal {
	if !utils.IsStrFilled(orgid) {
		orgid = pembuatSatusehat
	}
	return asal{orgid, pembuatSatusehat}
}

func asalPemanggil(orgid string, label *string) asal {
	if orgid == pembuatJakantro {
		return asalJakantro(label)
	}
	return asalFaskes(orgid)
}

func (a asal) orangtua(e *entity.Orangtua) {
	if e != nil {
		e.SourceData, e.CreatedBy = &a.sumber, &a.pembuat
	}
}

func (a asal) anak(e *entity.Anak) {
	if e != nil {
		e.SourceData, e.CreatedBy = a.sumber, &a.pembuat
	}
}

func (a asal) kunjungan(e *entity.Kunjungan) {
	if e != nil {
		e.SourceData, e.CreatedBy = &a.sumber, &a.pembuat
	}
}

func (a asal) kesehatan(e *entity.Kesehatan) {
	if e != nil {
		e.SourceData, e.CreatedBy = &a.sumber, &a.pembuat
	}
}

// label kiriman tiap entitas menjadi jakantro:<label>
func tandaiJakantro(o *entity.Orangtua, a *entity.Anak, k *entity.Kunjungan, h *entity.Kesehatan) {
	if o != nil {
		asalJakantro(o.SourceData).orangtua(o)
	}
	if a != nil {
		asalJakantro(&a.SourceData).anak(a)
	}
	if k != nil {
		asalJakantro(k.SourceData).kunjungan(k)
	}
	if h != nil {
		asalJakantro(h.SourceData).kesehatan(h)
	}
}
