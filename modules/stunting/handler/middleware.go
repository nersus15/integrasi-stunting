package handler

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
	"github.com/webcore-go/webcore/port/auth"
)

// GET: semua orgid. Selain GET: /api/faskes untuk faskes, sisanya jakantro
func Otorisasi(memory port.ICacheMemory) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == fiber.MethodGet {
			orgid, err := orgidDari(c, memory)
			if err != nil {
				return kirimError(c, err)
			}
			c.Locals("orgid", orgid)
			return c.Next()
		}

		role := "jakantro"
		if strings.HasPrefix(c.Path(), "/api/faskes") {
			role = "faskes"
		}
		if err := khusus(c, memory, role); err != nil {
			return kirimError(c, err)
		}
		return c.Next()
	}
}

func khusus(c *fiber.Ctx, memory port.ICacheMemory, role string) error {
	orgid, err := orgidDari(c, memory)

	if err != nil {
		return err
	}

	if orgid == nil || (role == "faskes" && utils.StrPtr(orgid) == "jakantro") || (role == "jakantro" && utils.StrPtr(orgid) != "jakantro") {
		return exceptions.Forbidden.Messagef("endpoint ini hanya untuk %s", role)
	}

	c.Locals("orgid", orgid)

	return nil
}

func orgidDari(c *fiber.Ctx, memory port.ICacheMemory) (*string, error) {
	var orgid string
	memorgid := new("")
	apikey := auth.GetAPIKey(c)
	memkey := "USERORGID::" + apikey

	if ok := memory.Get(memkey, memorgid); !ok {
		groups := auth.GetUserGroups(c)
		if groups == nil {
			return nil, exceptions.Forbidden.WithMessage("Gagal verifikasi hak akses user: user group tidak ditemukan", nil)
		}
		ada := false

		for _, g := range groups {
			if g == "jakantro" {
				ada = true
				orgid = g
				break
			} else if strings.Contains(g, "orgid:") {
				ada = true
				orgid = strings.Replace(g, "orgid:", "", 1)
				break
			}
		}

		if !ada {
			return nil, exceptions.Forbidden.WithMessage("Gagal verifikasi hak akses user", nil)
		}

		err := memory.Set(memkey, orgid, 24*time.Minute)

		if err != nil {
			logger.Error("CacheOrgid", err)
		}
	} else {
		logger.Info("Orgid dari cache: " + *memorgid)
		orgid = *memorgid
	}

	if !utils.IsStrFilled(orgid) {
		return nil, exceptions.Forbidden.WithMessage("Gagal verifikasi hak akses user: user group tidak ditemukan", nil)
	}

	return &orgid, nil
}
