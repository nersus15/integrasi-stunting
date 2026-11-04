package config

import (
	"time"

	"github.com/webcore-go/webcore/infra/config"
)

type ModuleConfig struct {
	Kafka        config.KafkaConfig `mapstructure:"kafka" json:"kafka"`
	Cron         CronConfig         `mapstructure:"cron" json:"cron"`
	LogTransaksi LogTransaksiConfig `mapstructure:"log_transaksi" json:"log_transaksi"`
}

var DefaultLogMask = []string{"nik", "no_kk", "nama", "nama_ayah", "nama_ibu", "telepon", "alamat"}

const (
	DefaultLogRedisKey = "stunting:log_transaksi"
	DefaultLogWorkers  = 10
)

type LogTransaksiConfig struct {
	Enabled       bool              `mapstructure:"enabled" json:"enabled"`
	Endpoint      string            `mapstructure:"endpoint" json:"endpoint"`
	Headers       map[string]string `mapstructure:"headers" json:"-"`
	RawHeaders    string            `mapstructure:"raw_headers" json:"-"`
	Header        string            `mapstructure:"header" json:"header"`
	Blacklist     []string          `mapstructure:"blacklist" json:"blacklist"`
	Mask          []string          `mapstructure:"mask" json:"mask"`
	Timeout       time.Duration     `mapstructure:"timeout" json:"timeout"`
	RedisKey      string            `mapstructure:"redis_key" json:"redis_key"`
	RetryInterval time.Duration     `mapstructure:"retry_interval" json:"retry_interval"`
	Workers       int               `mapstructure:"workers" json:"workers"`
}

type CronConfig struct {
	Enabled bool `mapstructure:"enabled" json:"enabled"`
}

func (c *ModuleConfig) SetEnvBindings() map[string]string {
	return map[string]string{
		"kafka.enabled": "KAFKA_ENABLED",
		"kafka.brokers": "KAFKA_BROKERS",
		"kafka.group":   "KAFKA_GROUP_ID",
		"kafka.offset":  "KAFKA_AUTO_OFFSET_RESET",
		"kafka.topics":  "KAFKA_TOPICS",

		"module.stunting.cron.enabled": "CRON_ENABLED",

		"module.stunting.log_transaksi.enabled":        "LOG_TRANSAKSI_ENABLED",
		"module.stunting.log_transaksi.endpoint":       "LOG_TRANSAKSI_ENDPOINT",
		"module.stunting.log_transaksi.header":         "LOG_TRANSAKSI_HEADER",
		"module.stunting.log_transaksi.blacklist":      "LOG_TRANSAKSI_BLACKLIST",
		"module.stunting.log_transaksi.mask":           "LOG_TRANSAKSI_MASK",
		"module.stunting.log_transaksi.timeout":        "LOG_TRANSAKSI_TIMEOUT",
		"module.stunting.log_transaksi.redis_key":      "LOG_TRANSAKSI_REDIS_KEY",
		"module.stunting.log_transaksi.retry_interval": "LOG_TRANSAKSI_RETRY_INTERVAL",
		"module.stunting.log_transaksi.workers":        "LOG_TRANSAKSI_WORKERS",
		"module.stunting.log_transaksi.raw_headers":    "LOG_TRANSAKSI_HEADERS",
	}
}

func (c *ModuleConfig) SetDefaults() map[string]any {
	data := map[string]any{
		"module.stunting.cron.enabled": true,

		"module.stunting.log_transaksi.enabled":        false,
		"module.stunting.log_transaksi.endpoint":       "",
		"module.stunting.log_transaksi.header":         "X-Transaction-ID",
		"module.stunting.log_transaksi.blacklist":      []string{},
		"module.stunting.log_transaksi.mask":           DefaultLogMask,
		"module.stunting.log_transaksi.timeout":        "5s",
		"module.stunting.log_transaksi.redis_key":      DefaultLogRedisKey,
		"module.stunting.log_transaksi.retry_interval": "10s",
		"module.stunting.log_transaksi.workers":        DefaultLogWorkers,
		"module.stunting.log_transaksi.raw_headers":    "",
	}

	return data
}
