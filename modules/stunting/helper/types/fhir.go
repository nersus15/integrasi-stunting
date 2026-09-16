package types

import "github.com/samply/golang-fhir-models/fhir-models/fhir"

type BundleEntryResponse struct {
	fhir.BundleEntryResponse

	// Override
	Id           *string `bson:"resourceID" json:"resourceID"`
	ResourceType *string `bson:"resourceType" json:"resourceType"`
}

type BundlePack struct {
	TransactionID *string               `json:"transactionId,omitempty"`
	Method        *string               `json:"method"`
	Env           *string               `json:"env"`
	Authorization *string               `json:"authorization,omitempty"`
	MetaProfile   *[]string             `json:"metaProfile"`
	Patient       *string               `json:"patient"`
	Input         *fhir.Bundle          `json:"input"`
	Response      []BundleEntryResponse `json:"response,omitempty"`
}
