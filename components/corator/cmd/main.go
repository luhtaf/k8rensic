package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/luhtaf/corator/config"
	"github.com/luhtaf/corator/detector"
	"github.com/luhtaf/corator/handler"
	"github.com/luhtaf/corator/logger"
	"github.com/luhtaf/corator/uploader"
	"github.com/luhtaf/corator/waf"
)

func main() {
	log.Println("Memulai Corator WAF Interceptor...")

	// 1. Muat Konfigurasi
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Gagal memuat konfigurasi: %v", err)
	}

	// 2. Inisialisasi semua komponen via factory
	loggers := logger.NewLoggers(&cfg)

	uploader, err := uploader.NewUploader(&cfg)
	if err != nil {
		log.Fatalf("Gagal membuat uploader: %v", err)
	}

	waf, err := waf.NewWAF(cfg.WAF)
	if err != nil {
		log.Fatalf("Gagal membuat WAF: %v", err)
	}

	detectors := detector.NewDetectors(&cfg)

	backendURL, err := url.Parse(cfg.Server.BackendURL)
	if err != nil {
		log.Fatalf("URL Backend tidak valid: %v", err)
	}

	// 3. Buat handler utama dan suntikkan semua komponen
	mainHandler := handler.NewRequestHandler(waf, detectors, uploader, loggers, backendURL)

	// 4. Jalankan server HTTP
	server := &http.Server{
		Addr:    cfg.Server.ListenAddress,
		Handler: mainHandler,
	}

	log.Printf("Server berjalan di %s", cfg.Server.ListenAddress)
	log.Printf("Meneruskan traffic ke backend: %s", backendURL)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server gagal berjalan: %v", err)
	}
}
