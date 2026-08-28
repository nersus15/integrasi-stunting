package config

import (
	"github.com/webcore-go/webcore/infra/config"
)

type ModuleConfig struct {
	Propagation            SatusehatPropagationConfig `mapstructure:"propagation" json:"propagation"`
	ResourceUseMasterToken []string                   `mapstructure:"use_master_token" json:"use_master_token"`
	FhirSource             FhirSource                 `mapstructure:"source" json:"source"`
	Kafka                  config.KafkaConfig         `mapstructure:"kafka" json:"kafka"`
	Cron                   CronConfig                 `mapstructure:"cron" json:"cron"`
	Ildki                  IldkiConfig                `mapstructure:"ildki" json:"ildki"`
	SatSetDev              SatusehatConfig            `mapstructure:"satusehat_dev" json:"satusehat_dev"`
	SatSetProd             SatusehatConfig            `mapstructure:"satusehat_prod" json:"satusehat_prod"`
}

type SatusehatPropagationConfig struct {
}

type SatusehatConfig struct {
	BaseURL      string `mapstructure:"baseurl" json:"baseurl"`
	AuthURL      string `mapstructure:"authurl" json:"authurl"`
	ClientID     string `mapstructure:"client_id" json:"client_id"`
	ClientSecret string `mapstructure:"client_secret" json:"client_secret"`
	HttpProxy    string `mapstructure:"http_proxy"`
}

type IldkiConfig struct {
	Faskes         string `mapstructure:"faskes_id" json:"faskes_id"`
	Enabled        bool   `mapstructure:"enabled" json:"enabled"`
	ProductionURL  string `mapstructure:"production_url" json:"production_url"`
	DevelopmentURL string `mapstructure:"development_url" json:"development_url"`
	DevAuthUrl     string `mapstructure:"development_authurl" json:"development_authurl"`
	ProdAuthUrl    string `mapstructure:"production_authurl" json:"production_authurl"`
	BackupBaseUrl  string `mapstructure:"backup_baseurl" json:"backup_baseurl"`
	ZeroTrust      bool   `mapstructure:"zero_trust" json:"zero_trust"`
	HttpProxy      string `mapstructure:"http_proxy" json:"http_proxy"`
}

type FhirSource struct {
	Priority string `mapstructure:"priority" json:"priority"`
	Header   string `mapstructure:"header" json:"header"`
}

type CronConfig struct {
	Enabled     bool   `mapstructure:"enabled" json:"enabled"`
	Schedule    string `mapstructure:"schedule" json:"schedule"`
	ChunkSize   int64  `mapstructure:"backup_chunksize" json:"backup_chunksize"`
	BackupLimit int64  `mapstructure:"backup_limit" json:"backup_limit"`
}

func (c *ModuleConfig) SetEnvBindings() map[string]string {
	return map[string]string{
		"module.fhir.ildki.enabled":             "MODULE_FHIR_ILDKI_ENABLED",
		"module.fhir.ildki.production_url":      "MODULE_FHIR_ILDKI_PRODUCTION_URL",
		"module.fhir.ildki.development_url":     "MODULE_FHIR_ILDKI_DEVELOPMENT_URL",
		"module.fhir.ildki.development_authurl": "MODULE_FHIR_ILDKI_DEVELOPMENT_AUTHURL",
		"module.fhir.ildki.production_authurl":  "MODULE_FHIR_ILDKI_PRODUCTION_AUTHURL",
		"module.fhir.ildki.zero_trust":          "MODULE_FHIR_ILDKI_ZERO_TRUST",
		"module.fhir.ildki.backup_baseurl":      "MODULE_FHIR_ILDKI_BACKUP_BASEURL",

		"module.fhir.satusehat_prod.baseurl":       "MODULE_FHIR_PRODUCTION_BASEURL",
		"module.fhir.satusehat_prod.authurl":       "MODULE_FHIR_PRODUCTION_AUTHURL",
		"module.fhir.satusehat_prod.client_id":     "MODULE_FHIR_PRODUCTION_CLIENT_ID",
		"module.fhir.satusehat_prod.client_secret": "MODULE_FHIR_PRODUCTION_CLIENT_SECRET",
		"module.fhir.satusehat_prod.http_proxy":    "MODULE_FHIR_PRODUCTION_HTTP_PROXY",

		"module.fhir.satusehat_dev.baseurl":       "MODULE_FHIR_DEVELOPMENT_BASEURL",
		"module.fhir.satusehat_dev.authurl":       "MODULE_FHIR_DEVELOPMENT_AUTHURL",
		"module.fhir.satusehat_dev.client_id":     "MODULE_FHIR_DEVELOPMENT_CLIENT_ID",
		"module.fhir.satusehat_dev.client_secret": "MODULE_FHIR_DEVELOPMENT_CLIENT_SECRET",
		"module.fhir.satusehat_dev.http_proxy":    "MODULE_FHIR_DEVELOPMENT_HTTP_PROXY",

		"module.fhir.use_master_token": "MODULE_FHIR_RESOURCE_USE_MASTER_TOKEN",

		"module.fhir.propagation.auto":   "MODULE_FHIR_PROPAGATION_AUTO",
		"module.fhir.propagation.header": "MODULE_FHIR_PROPAGATION_HEADER",

		"module.fhir.source.priority": "MODULE_FHIR_SOURCE_PRIORITY",
		"module.fhir.source.header":   "MODULE_FHIR_SOURCE_PRIORITY_HEADER",

		// Kafka
		"kafka.enabled": "KAFKA_ENABLED",
		"kafka.brokers": "KAFKA_BROKERS",
		"kafka.group":   "KAFKA_GROUP_ID",
		"kafka.offset":  "KAFKA_AUTO_OFFSET_RESET",
		"kafka.topics":  "KAFKA_TOPICS",

		// CRON
		"module.fhir.cron.enabled":          "CRON_ENABLED",
		"module.fhir.cron.schedule":         "CRON_SCHEDULE",
		"module.fhir.cron.backup_chunksize": "CRON_BACKUP_CHUNKSIZE",
		"module.fhir.cron.backup_limit":     "CRON_BACKUP_LIMIT",
	}
}

func (c *ModuleConfig) SetDefaults() map[string]any {
	data := map[string]any{
		"module.fhir.ildki.enabled":             true,
		"module.fhir.ildki.production_url":      "https://api.ildki.appgo.my.id/prod/fhir/r4/v1",
		"module.fhir.ildki.development_url":     "https://api.ildki.appgo.my.id/dev/fhir/r4/v1",
		"module.fhir.ildki.development_authurl": "https://api.ildki.appgo.my.id/dev/oauth2/v1/accesstoken?grant_type=client_credentials",
		"module.fhir.ildki.production_authurl":  "https://api.ildki.appgo.my.id/prod/oauth2/v1/accesstoken?grant_type=client_credentials",
		"module.fhir.ildki.zero_trust":          false,
		"module.fhir.ildki.backup_baseurl":      "",

		"module.fhir.satusehat_prod.baseurl":       "https://api-satusehat.kemkes.go.id/fhir-r4/v1",
		"module.fhir.satusehat_prod.authurl":       "https://api-satusehat.kemkes.go.id/oauth2/v1/accesstoken?grant_type=client_credentials",
		"module.fhir.satusehat_prod.client_id":     "your-production-client-id",
		"module.fhir.satusehat_prod.client_secret": "your-production-client-secret",

		"module.fhir.satusehat_dev.baseurl":       "https://api-satusehat-stg.dto.kemkes.go.id/fhir-r4/v1",
		"module.fhir.satusehat_dev.authurl":       "https://api-satusehat-stg.dto.kemkes.go.id/oauth2/v1/accesstoken?grant_type=client_credentials",
		"module.fhir.satusehat_dev.client_id":     "your-development-client-id",
		"module.fhir.satusehat_dev.client_secret": "your-development-client-secret",

		"module.fhir.use_master_token": []string{},

		"module.fhir.propagation.auto":   false,
		"module.fhir.propagation.header": "X-Propagation-Strategy", // pilihan ['satusehat-failover', 'local-only']

		"module.fhir.source.priority":       "satusehat-first",
		"module.fhir.source.header":         "X-FHIR-Source-Priority",
		"module.fhir.cron.enabled":          true,
		"module.fhir.cron.schedule":         "*/10 23 * * *",
		"module.fhir.cron.backup_chunksize": 100,
		"module.fhir.cron.backup_limit":     5,
	}

	// tambahkan key hapi dengan value sama persis dari ildki
	data["module.fhir.hapi.enabled"] = data["module.fhir.ildki.enabled"]
	data["module.fhir.hapi.production_url"] = data["module.fhir.ildki.production_url"]
	data["module.fhir.hapi.development_url"] = data["module.fhir.ildki.development_url"]
	data["module.fhir.hapi.production_authurl"] = data["module.fhir.ildki.production_authurl"]
	data["module.fhir.hapi.development_authurl"] = data["module.fhir.ildki.development_authurl"]

	return data
}
