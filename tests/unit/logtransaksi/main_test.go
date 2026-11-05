package logtransaksi_test

import (
	"context"
	"os"
	"testing"

	"github.com/webcore-go/webcore/infra/logger"
)

func TestMain(m *testing.M) {
	logger.PrepareLogger(context.Background(), "error")
	os.Exit(m.Run())
}
