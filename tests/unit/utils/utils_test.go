package utils_test

import (
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
)

// --- POSITIVE CASE ---

func TestIsFilled_PointerBerisi(t *testing.T) {
	assert.True(t, utils.IsFilled(testutil.Ptr("nilai")))
	assert.True(t, utils.IsFilled(testutil.Ptr("0")))
}

func TestIsStrFilled_StringBerisi(t *testing.T) {
	assert.True(t, utils.IsStrFilled("nilai"))
	assert.True(t, utils.IsStrFilled(" a "))
}

func TestIsValidDate_FormatBenar(t *testing.T) {
	for _, d := range []string{"2025-02-16", "2024-05-10", "2000-01-01"} {
		assert.True(t, utils.IsValidDate(d), d)
	}
}

// --- NEGATIVE CASE ---

func TestIsFilled_PointerKosong(t *testing.T) {
	assert.False(t, utils.IsFilled(nil))
	assert.False(t, utils.IsFilled(testutil.Ptr("")))
	assert.False(t, utils.IsFilled(testutil.Ptr("   ")))
	assert.False(t, utils.IsFilled(testutil.Ptr("\t\n")))
}

func TestIsStrFilled_StringKosong(t *testing.T) {
	assert.False(t, utils.IsStrFilled(""))
	assert.False(t, utils.IsStrFilled("   "))
	assert.False(t, utils.IsStrFilled("\t\n"))
}

func TestIsValidDate_FormatSalah(t *testing.T) {
	tests := []string{
		"", "16-02-2025", "2025/02/16", "2025-2-16",
		"kemarin", "2025-13-01", "2025-02-30", "2025-02-16T00:00:00Z",
	}

	for _, d := range tests {
		t.Run(d, func(t *testing.T) {
			assert.False(t, utils.IsValidDate(d))
		})
	}
}
