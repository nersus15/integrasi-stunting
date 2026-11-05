package create_orangtua_anak_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Body JSON literal `null` adalah kasus khusus: untuk struct value ia
// meninggalkan zero value, untuk pointer ia meninggalkan nil.
func TestBodyNullDitolak(t *testing.T) {
	for _, path := range []string{pathOrangtua, pathAnak} {
		t.Run(path, func(t *testing.T) {
			resp := post(t, path, []byte(`null`))
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode,
				"body null tidak membawa data apa pun, harus ditolak")
		})
	}
}
