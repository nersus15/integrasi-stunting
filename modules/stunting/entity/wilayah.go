package entity

import (
	"strings"
)

// level wilayah => [0 => Nasional, 1 => Provinsi, 2 => Kab/Kota, 3 => Kecamatan, 4 => Kelurahan, -1 => Invalid]
func LevelWilayah(kodeWilayah *string) int {
	if kodeWilayah == nil || strings.TrimSpace(*kodeWilayah) == "" {
		return -1
	}
	arr := strings.Split(*kodeWilayah, ".")
	level := 0

	for _, s := range arr {
		if s == "00" || s == "0000" {
			continue
		}
		level += 1
	}

	return level
}
