package satusehat

import (
	"encoding/json"
	"sort"
	"strings"

	types2 "github.com/semanggilab/lib-go-fhir/helper/types"
)

var BobotResourceBundle = map[string]int{
	// prerequisite / master
	"organization":      10,
	"location":          11,
	"healthcareservice": 12,
	"practitioner":      13,
	"practitionerrole":  14,
	"patient":           15,
	"relatedperson":     16,
	"coverage":          17,

	// konteks kunjungan
	"episodeofcare":       20,
	"slot":                21,
	"appointment":         22,
	"appointmentresponse": 23,
	"encounter":           24,

	// data mentah, hanya butuh patient + encounter
	"allergyintolerance":  30,
	"familymemberhistory": 31,
	"medicationstatement": 32,
	"immunization":        33,
	"specimen":            34,
	"observation":         35,
	"imagingstudy":        36,

	// turunan dari data mentah
	"diagnosticreport":      40,
	"questionnaireresponse": 41,

	// kesimpulan klinis
	"condition":          50,
	"clinicalimpression": 51,
	"riskassessment":     52,
	"goal":               53,

	// intervensi
	"procedure":                60,
	"medication":               61,
	"medicationrequest":        62,
	"medicationadministration": 63,
	"medicationdispense":       64,
	"nutritionorder":           65,
	"servicerequest":           66,
	"careplan":                 67,
	"task":                     68,

	// dokumen
	"composition":       70,
	"documentreference": 71,

	// pembiayaan
	"account":                     80,
	"chargeitem":                  81,
	"claim":                       82,
	"claimresponse":               83,
	"invoice":                     84,
	"paymentnotice":               85,
	"paymentreconciliation":       86,
	"coverageeligibilityrequest":  87,
	"coverageeligibilityresponse": 88,
}

func BobotEntry(raw json.RawMessage) int {
	var res struct {
		ResourceType string          `json:"resourceType"`
		DerivedFrom  json.RawMessage `json:"derivedFrom"`
	}

	if err := json.Unmarshal(raw, &res); err != nil {
		return -1
	}

	bobot := BobotResource(res.ResourceType)

	if strings.EqualFold(res.ResourceType, "Observation") {
		var d []json.RawMessage
		if json.Unmarshal(res.DerivedFrom, &d) == nil && len(d) > 0 {
			bobot++
		}
	}
	return bobot
}

// bobot berpasangan dengan entry-nya supaya cukup dihitung sekali
type urutBobot struct {
	entry []types2.BundleEntry
	bobot []int
}

func (u *urutBobot) Len() int           { return len(u.entry) }
func (u *urutBobot) Less(i, j int) bool { return u.bobot[i] < u.bobot[j] }
func (u *urutBobot) Swap(i, j int) {
	u.entry[i], u.entry[j] = u.entry[j], u.entry[i]
	u.bobot[i], u.bobot[j] = u.bobot[j], u.bobot[i]
}

func UrutkanEntry(entries []types2.BundleEntry) {
	bobot := make([]int, len(entries))
	for i := range entries {
		bobot[i] = BobotEntry(entries[i].Resource)
	}
	sort.Stable(&urutBobot{entry: entries, bobot: bobot})
}

const BobotDefault = 999

func BobotResource(resourceType string) int {
	if b, ok := BobotResourceBundle[strings.ToLower(resourceType)]; ok {
		return b
	}
	return BobotDefault
}
