package detector

import (
	"log"

	"github.com/luhtaf/corator/config"
)

// NewDetectors adalah factory yang membuat semua detektor yang aktif.
func NewDetectors(cfg *config.Config) []Detector {
	var activeDetectors []Detector

	if cfg.Detectors.EnableFile {
		log.Println("FileDetector aktif.")
		activeDetectors = append(activeDetectors, NewFileDetector())
	}

	if cfg.Detectors.EnableBase64 {
		log.Println("Base64Detector aktif.")
		activeDetectors = append(activeDetectors, NewBase64Detector())
	}

	return activeDetectors
}
