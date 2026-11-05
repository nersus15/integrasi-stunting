package utils

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/tink-crypto/tink-go/v2/daead/subtle"
)

type Field string

const (
	FieldNik  Field = "nik"
	FieldNoKK Field = "no_kk"
)

// versi chipertext untuk rotasi key
const prefix = "e1:"

var ErrDecrypt = errors.New("AES-SIV: gagal dekripsi")

type Crypter struct {
	siv *subtle.AESSIV
}

func NewCrypter(keyBase64 string) (*Crypter, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(keyBase64))
	if err != nil {
		return nil, err
	}

	if len(key) != subtle.AESSIVKeySize {
		return nil, fmt.Errorf("AES-SIV: key harus %d byte, dapat %d", subtle.AESSIVKeySize, len(key))
	}
	siv, err := subtle.NewAESSIV(key)
	if err != nil {
		return nil, err
	}
	return &Crypter{siv: siv}, nil
}

func (c *Crypter) Seal(f Field, v string) (string, error) {
	ct, err := c.siv.EncryptDeterministically([]byte(v), []byte(f))
	if err != nil {
		return "", err
	}
	return prefix + base64.RawURLEncoding.EncodeToString(ct), nil
}

func (c *Crypter) Open(f Field, s string) (string, error) {
	if !strings.HasPrefix(s, prefix) {
		return s, nil
	}
	ct, err := base64.RawURLEncoding.DecodeString(s[len(prefix):])
	if err != nil {
		return "", ErrDecrypt
	}
	pt, err := c.siv.DecryptDeterministically(ct, []byte(f))
	if err != nil {
		return "", ErrDecrypt
	}
	return string(pt), nil
}
