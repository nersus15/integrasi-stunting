package satusehat

import (
	satusehat "github.com/nersus15/integrasi/mod-satusehat"

	"github.com/webcore-go/webcore/app/core"
)

var APP_PACKAGES = []core.Module{
	satusehat.NewModule(),
}
