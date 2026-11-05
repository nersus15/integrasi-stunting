package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type Response struct {
	Status    int
	Body      map[string]any
	ErrorCode int
}

func Request(t *testing.T, url string, method, key string, body any) Response {
	t.Helper()
	var r io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		r = bytes.NewBufferString(b)
	default:
		content, err := json.Marshal(b)
		require.NoError(t, err)
		r = bytes.NewReader(content)
	}

	req, err := http.NewRequest(method, url, r)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("X-API-Key", key)
	}
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	j := Response{Status: res.StatusCode}
	content, _ := io.ReadAll(res.Body)
	_ = json.Unmarshal(content, &j.Body)
	if k, ok := j.Body["errorCode"].(float64); ok {
		j.ErrorCode = int(k)
	}
	return j
}
