package app

import (
	stunting "github.com/nersus15/integrasi/mod-stunting"
	"github.com/webcore-go/webcore/app/core"
)

var APP_PACKAGES = []core.Module{
	stunting.NewModule(),
	// Add your packages here
}
