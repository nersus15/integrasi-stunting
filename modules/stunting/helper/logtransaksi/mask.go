package logtransaksi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Record struct {
	TransactionID string
	Timestamp     time.Time
	Duration      time.Duration
	Method        string
	Path          string
	Query         string
	Status        int
	Orgid         string
	IP            string
	UserAgent     string
	Request       []byte
	Response      []byte
	Secrets       []string
}

type payload struct {
	TransactionID string  `json:"transaction_id"`
	Timestamp     string  `json:"timestamp"`
	DurationMs    float64 `json:"duration_ms"`
	Method        string  `json:"method"`
	Path          string  `json:"path"`
	Query         string  `json:"query,omitempty"`
	Status        int     `json:"status"`
	Orgid         string  `json:"orgid,omitempty"`
	IP            string  `json:"ip,omitempty"`
	UserAgent     string  `json:"user_agent,omitempty"`
	Request       any     `json:"request,omitempty"`
	Response      any     `json:"response,omitempty"`
}

type envelope struct {
	Level      string  `json:"level"`
	Message    string  `json:"message"`
	Identifier string  `json:"identifier"`
	Context    payload `json:"context"`
}

type Masker struct {
	fields map[string]bool
}

func NewMasker(fields []string) *Masker {
	m := &Masker{fields: map[string]bool{}}
	for _, f := range fields {
		m.fields[strings.ToLower(f)] = true
	}
	return m
}

func (m *Masker) Prepare(r Record) any {
	p := payload{
		TransactionID: r.TransactionID,
		Timestamp:     r.Timestamp.Format(time.RFC3339Nano),
		DurationMs:    float64(r.Duration.Microseconds()) / 1000,
		Method:        r.Method,
		Path:          r.Path,
		Query:         m.maskQuery(string(maskSecrets([]byte(r.Query), r.Secrets))),
		Status:        r.Status,
		Orgid:         r.Orgid,
		IP:            r.IP,
		UserAgent:     r.UserAgent,
		Request:       m.maskBody(maskSecrets(r.Request, r.Secrets)),
		Response:      m.maskBody(maskSecrets(r.Response, r.Secrets)),
	}
	return envelope{Level: "info", Message: responseMessage(p.Response, r.Status), Identifier: r.TransactionID, Context: p}
}

func responseMessage(res any, status int) string {
	if obj, ok := res.(map[string]any); ok {
		if msg, ok := obj["message"].(string); ok && msg != "" {
			return msg
		}
	}
	if status >= 200 && status < 300 {
		return "Berhasil"
	}
	if text := http.StatusText(status); text != "" {
		return text
	}
	return fmt.Sprintf("Status %d", status)
}

const minSecretLen = 8

func maskSecrets(raw []byte, secrets []string) []byte {
	for _, s := range secrets {
		if len(s) >= minSecretLen && bytes.Contains(raw, []byte(s)) {
			raw = bytes.ReplaceAll(raw, []byte(s), []byte(maskText(s)))
		}
	}
	return raw
}

func (m *Masker) maskBody(raw []byte) any {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}

	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var v any
	if err := d.Decode(&v); err == nil && !d.More() {
		return m.maskJSON(v)
	}
	return longDigits.ReplaceAllStringFunc(string(raw), maskText)
}

func (m *Masker) maskJSON(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if m.fields[strings.ToLower(k)] {
				switch s := val.(type) {
				case string:
					x[k] = maskText(s)
					continue
				case json.Number:
					x[k] = maskText(s.String())
					continue
				}
			}
			x[k] = m.maskJSON(val)
		}
	case []any:
		for i := range x {
			x[i] = m.maskJSON(x[i])
		}
	}
	return v
}

func (m *Masker) maskQuery(q string) string {
	if q == "" {
		return ""
	}
	values, err := url.ParseQuery(q)
	if err != nil {
		return longDigits.ReplaceAllStringFunc(q, maskText)
	}
	for k, vals := range values {
		if m.fields[strings.ToLower(k)] {
			for i := range vals {
				vals[i] = maskText(vals[i])
			}
		}
	}
	return values.Encode()
}

var longDigits = regexp.MustCompile(`\d{8,}`)

// karakter terakhir yang dibiarkan supaya data masih bisa dikenali
const visibleTail = 3

func maskText(s string) string {
	r := []rune(s)
	if len(r) <= visibleTail {
		return strings.Repeat("*", len(r))
	}
	return strings.Repeat("*", len(r)-visibleTail) + string(r[len(r)-visibleTail:])
}
