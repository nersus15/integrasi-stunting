package ratelimit_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
)

func TestMain(m *testing.M) {
	testutil.SetupLogger()
	os.Exit(m.Run())
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
