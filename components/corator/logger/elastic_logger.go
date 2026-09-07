package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/luhtaf/corator/config"
)

// ElasticLogger adalah implementasi logger yang mengirim log ke Elasticsearch.
type ElasticLogger struct {
	client *elasticsearch.Client
	index  string
}

// NewElasticLogger membuat instance baru dari ElasticLogger.
func NewElasticLogger(cfg config.ElasticConfig) (*ElasticLogger, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.URLs,
	}
	esClient, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, err
	}
	return &ElasticLogger{
		client: esClient,
		index:  cfg.Index,
	}, nil
}

// Log mengirimkan event ke Elasticsearch secara asinkron.
func (l *ElasticLogger) Log(event LogEvent) {
	// Jalankan dalam goroutine agar tidak memblokir proses utama
	go func() {
		body, err := json.Marshal(event)
		if err != nil {
			log.Printf("ElasticLogger: Gagal marshal log event: %v", err)
			return
		}

		_, err = l.client.Index(
			l.index,
			bytes.NewReader(body),
			l.client.Index.WithContext(context.Background()),
		)
		if err != nil {
			log.Printf("ElasticLogger: Gagal mengirim log: %v", err)
		}
	}()
}
