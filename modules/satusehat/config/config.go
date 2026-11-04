// Package config berisi konfigurasi khusus modul satusehat, dibaca dari
// namespace "module.satusehat.*" di config.yaml / env MODULE_SATUSEHAT_*.
package config

type ModuleConfig struct {
	PageSize int `mapstructure:"page_size"`
}

func (c *ModuleConfig) SetDefaults() map[string]any {
	return map[string]any{
		"page_size": 20,
	}
}

func (c *ModuleConfig) SetEnvBindings() map[string]string {
	return map[string]string{
		"page_size": "MODULE_SATUSEHAT_PAGE_SIZE",
	}
}
