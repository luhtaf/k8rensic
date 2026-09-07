package logger

import (
	"log"

	"github.com/luhtaf/corator/config"
)

// NewLoggers adalah factory yang membuat semua logger yang aktif berdasarkan konfigurasi.
func NewLoggers(cfg *config.Config) []Logger {
	var activeLoggers []Logger

	if cfg.Logger.EnableFile {
		fileLogger, err := NewFileLogger(cfg.Logger.File)
		if err != nil {
			log.Printf("PERINGATAN: Gagal menginisialisasi FileLogger: %v", err)
		} else {
			log.Println("FileLogger aktif.")
			activeLoggers = append(activeLoggers, fileLogger)
		}
	}

	if cfg.Logger.EnableElastic {
		elasticLogger, err := NewElasticLogger(cfg.Logger.Elastic)
		if err != nil {
			log.Printf("PERINGATAN: Gagal menginisialisasi ElasticLogger: %v", err)
		} else {
			log.Println("ElasticLogger aktif.")
			activeLoggers = append(activeLoggers, elasticLogger)
		}
	}

	return activeLoggers
}
