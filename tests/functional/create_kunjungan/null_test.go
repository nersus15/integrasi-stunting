package create_kunjungan_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Di alur kunjungan payload-nya pointer, jadi body `null` menghasilkan nil -
// bukan zero value seperti pada orangtua dan anak.
func TestBodyNullDitolak(t *testing.T) {
	resp := post(t, path, []byte(`null`))
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode,
		"body null harus ditolak, bukan membuat service men-dereference payload nil")
}
