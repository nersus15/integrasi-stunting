package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	types2 "github.com/semanggilab/lib-go-fhir/helper/types"
	"github.com/semanggilab/lib-go-fhir/processor"
	"github.com/webcore-go/webcore/app/helper"
	"github.com/webcore-go/webcore/infra/logger"
)

func ekstrakPatientId(message []byte) *string {
	tx := types2.PackMediator{}
	if err := json.Unmarshal(message, &tx); err != nil {
		return nil
	}

	if !utils.IsFilled(tx.Patient) {
		return nil
	}

	return tx.Patient
}

func (s *StreamService) extractRequestAndResponse(tx types2.PackMediator) (*fhir.Bundle, []types2.BundleEntryResponse, error) {
	var request *fhir.Bundle
	var response []types2.BundleEntryResponse

	if tx.Input != nil {
		request = tx.Input
	} else {
		return nil, nil, fmt.Errorf("request body kosong")
	}

	// response bisa Bundle atau OperationOutcome, hanya Bundle yang diproses
	if tx.Response != nil {
		response = tx.Response
		if len(request.Entry) != len(response) {
			return nil, nil, fmt.Errorf("jumlah entry Bundle request dan response tidak sama")
		}
	} else {
		return nil, nil, fmt.Errorf("response body kosong")
	}

	return request, response, nil
}

func (s *StreamService) SusunUlangBundle(tx types2.PackMediator) (*types2.Bundle, *types2.SetReference, error) {
	requestBundle, response, err := s.extractRequestAndResponse(tx)

	if err != nil {
		return nil, nil, err
	}
	processor.NormalizeTemporaryReferences(requestBundle)

	if utils.DumpAktif() {
		logger.Debug("Request Bundle Setelah Normalisasi: " + helper.ToLogJSON(requestBundle))
	}

	var register types2.SetReference = make(types2.SetReference)
	newBundle := types2.CastToBundle(*requestBundle)
	idBaru := []types2.NewPost{}
	for i := range requestBundle.Entry {
		if response[i].Id == nil || *response[i].Id == "" {
			return nil, nil, fmt.Errorf("entry ke-%d tidak punya resourceID di response", i)
		}

		responseId := response[i].Id
		entry, err := processor.NewBundleEntry(requestBundle.Entry[i], responseId)
		if entry == nil || err != nil {
			continue
		}

		idBaru = s.mapAllReferences(&register, idBaru, entry, tx.Response[i], i, 0)

		newBundle.Entry[i] = *entry

	}
	if len(idBaru) > 0 {
		for i := range newBundle.Entry {
			if newBundle.Entry[i].Base == nil {
				continue
			}
			processor.UpdateTemporaryReference(&newBundle.Entry[i], &register, idBaru)
		}
	}
	return &newBundle, &register, err
}

func (h *StreamService) mapAllReferences(register *types2.SetReference, baru []types2.NewPost, entry *types2.BundleEntry, response types2.BundleEntryResponse, i int, priority int) []types2.NewPost {
	if response.Id == nil {
		return baru
	}

	id := response.Id

	urlUrn := entry.GetTemporaryURN()
	if urlUrn != nil {
		if _, ok := (*register)[*id]; !ok {
			urn := (*urlUrn)[9:]
			sre := &types2.SetReferenceEntry{
				Index:              i,
				Priority:           priority,
				Temporary:          true,
				ReferredByTemp:     []string{},
				ResourceType:       *response.ResourceType,
				URN:                &urn,
				Id:                 id,
				Entry:              entry,
				IsSatusehatChecked: true,
				InSatusehat:        types2.DONE,
			}

			(*register)[*id] = sre

			baru = append(baru, types2.NewPost{
				Index:        i,
				ResourceType: *entry.Base.ResourceType,
				NewID:        *id,
				TempID:       urn,
			})
		}
	} else {
		sre, ok := (*register)[*id]
		if !ok {
			sre = &types2.SetReferenceEntry{
				Index:              i,
				Priority:           priority,
				ResourceType:       *response.ResourceType,
				Id:                 id,
				Entry:              entry,
				IsSatusehatChecked: true,
				InSatusehat:        types2.DONE,
			}

			(*register)[*id] = sre
		} else {
			if priority > sre.Priority {
				sre.Priority = priority
			}

			if !sre.IsSatusehatChecked {
				sre.IsSatusehatChecked = true
				sre.InSatusehat = types2.DONE
			}
		}
	}

	references := processor.ExtractReferences(*entry.Base)

	for _, r := range references {
		if r.Reference == nil || *r.Reference == "" {
			continue
		}

		// temporary id ditangani di level resource
		rp := strings.Split(*r.Reference, "/")
		if len(rp) == 2 {
			restype := rp[0]
			id := rp[1]
			reg, ok := (*register)[id]
			newPriority := priority + 1

			if !ok {
				sre := &types2.SetReferenceEntry{
					Index:              -1, // resource di luar bundle
					Priority:           newPriority,
					ResourceType:       restype,
					Id:                 &id,
					IsSatusehatChecked: true,
					InSatusehat:        types2.DONE,
				}
				(*register)[id] = sre
			} else {
				if reg.Priority < newPriority {
					(*register)[id].Priority = newPriority
				}
			}
		}
	}

	return baru
}
