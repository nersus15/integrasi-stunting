package handler

import (
	"bytes"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/nersus15/integrasi/mod-stunting/helper/logtransaksi"
)

var validTransactionID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

func LogTransaksi(cfg config.LogTransaksiConfig, apiKeyHeader string, push func(logtransaksi.Record)) fiber.Handler {
	header := cfg.Header
	if header == "" {
		header = "X-Transaction-ID"
	}
	return func(c *fiber.Ctx) error {
		id := strings.Clone(c.Get(header))
		if !validTransactionID.MatchString(id) {
			id = uuid.NewString()
		}
		c.Set(header, id)
		c.Locals("transaction_id", id)

		if isBlacklisted(cfg.Blacklist, c.Method(), c.Path()) {
			return c.Next()
		}

		start := time.Now()
		record := logtransaksi.Record{
			TransactionID: id,
			Timestamp:     start,
			Method:        strings.Clone(c.Method()),
			Path:          strings.Clone(c.Path()),
			Query:         string(c.Request().URI().QueryString()),
			IP:            strings.Clone(c.IP()),
			UserAgent:     strings.Clone(c.Get(fiber.HeaderUserAgent)),
			Request:       bytes.Clone(c.Body()),
			Secrets:       credentials(c, apiKeyHeader),
		}

		defer func() {
			if r := recover(); r != nil {
				record.Duration, record.Status, record.Orgid = time.Since(start), fiber.StatusInternalServerError, orgidPemanggil(c)
				push(record)
				panic(r)
			}
		}()

		err := c.Next()

		record.Duration = time.Since(start)
		record.Status = c.Response().StatusCode()
		record.Orgid = orgidPemanggil(c)
		record.Response = bytes.Clone(c.Response().Body())
		push(record)
		return err
	}
}

func credentials(c *fiber.Ctx, apiKeyHeader string) []string {
	var res []string
	if apiKeyHeader != "" {
		if k := c.Get(apiKeyHeader); k != "" {
			res = append(res, strings.Clone(k))
		}
	}
	if a := c.Get(fiber.HeaderAuthorization); a != "" {
		_, token, ok := strings.Cut(a, " ")
		if !ok {
			token = a
		}
		res = append(res, strings.Clone(strings.TrimSpace(token)))
	}
	return res
}

func isBlacklisted(blacklist []string, method, path string) bool {
	for _, pattern := range blacklist {
		pattern = strings.TrimSpace(pattern)
		if m, rest, ok := strings.Cut(pattern, " "); ok {
			if !strings.EqualFold(m, method) {
				continue
			}
			pattern = strings.TrimSpace(rest)
		}
		if prefix, ok := strings.CutSuffix(pattern, "*"); ok {
			if strings.HasPrefix(path, prefix) {
				return true
			}
		} else if path == pattern {
			return true
		}
	}
	return false
}
