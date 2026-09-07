package waf

import (
	"fmt"

	"github.com/corazawaf/coraza/v3"

	// KOREKSI: Import "github.com/corazawaf/coraza/v3/seclang" DIHAPUS
	"github.com/luhtaf/corator/config"
)

// NewWAF menginisialisasi WAF engine Coraza dari path konfigurasi.
func NewWAF(cfg config.WAFConfig) (coraza.WAF, error) {
	if cfg.CorazaConfigPath == "" {
		return nil, fmt.Errorf("path konfigurasi Coraza (WAF_CORAZA_CONFIG_PATH) tidak boleh kosong")
	}

	waf, err := coraza.NewWAF(
		coraza.NewWAFConfig().
			WithDirectivesFromFile(cfg.CorazaConfigPath).
			WithRequestBodyAccess().
			WithResponseBodyAccess(),
	)

	if err != nil {
		// KOREKSI: Pengecekan error spesifik ke seclang.ParserError DIHAPUS.
		// Pesan error umum dari Coraza sudah cukup jelas.
		return nil, fmt.Errorf("gagal membuat instance WAF Coraza: %w", err)
	}

	return waf, nil
}
