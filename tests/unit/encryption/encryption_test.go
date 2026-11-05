package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/stretchr/testify/assert"
)

func GenerateSecret(size int) (string, error) {
	b := make([]byte, size)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}
func TestDeterministicEncrypt(t *testing.T) {
	key, err := GenerateSecret(64)
	nik := "5203192408990002"

	if err != nil {
		fmt.Printf("Gagal membuat key")
	}

	enc, err := utils.NewCrypter(key)

	if err != nil {
		fmt.Printf("Gagal inisiasi function")
	}

	hashed, err := enc.Seal(utils.FieldNik, nik)

	if err != nil {
		fmt.Printf("Gagal hash data")
	}

	hashed2, err := enc.Seal(utils.FieldNik, nik)
	if err != nil {
		fmt.Printf("Gagal hash data")
	}

	// hashed dan hashed2 harus sama karena determenistic
	assert.Equal(t, hashed, hashed2, "%s dan %s tidak sama", hashed, hashed2)
}
