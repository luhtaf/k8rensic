package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config menampung semua konfigurasi untuk aplikasi.
// Nilai-nilai ini dibaca dari environment variables.
type Config struct {
	Server    ServerConfig
	WAF       WAFConfig
	Detectors DetectorConfig
	Uploader  UploaderConfig
	Logger    LoggerConfig
}

type ServerConfig struct {
	ListenAddress string `mapstructure:"LISTEN_ADDRESS"`
	BackendURL    string `mapstructure:"BACKEND_URL"`
}

type WAFConfig struct {
	CorazaConfigPath string `mapstructure:"CORAZA_CONFIG_PATH"`
}

type DetectorConfig struct {
	EnableFile   bool `mapstructure:"ENABLE_FILE_DETECTOR"`
	EnableBase64 bool `mapstructure:"ENABLE_BASE64_DETECTOR"`
}

type UploaderConfig struct {
	Type  string      `mapstructure:"TYPE"` // "local" atau "s3"
	Local LocalConfig `mapstructure:"LOCAL"`
	S3    S3Config    `mapstructure:"S3"`
}

type LocalConfig struct {
	Path string `mapstructure:"PATH"`
}

type S3Config struct {
	Endpoint  string `mapstructure:"ENDPOINT"`
	Bucket    string `mapstructure:"BUCKET"`
	Region    string `mapstructure:"REGION"`
	AccessKey string `mapstructure:"ACCESS_KEY"`
	SecretKey string `mapstructure:"SECRET_KEY"`
}

type LoggerConfig struct {
	EnableFile    bool             `mapstructure:"ENABLE_FILE"`
	EnableElastic bool             `mapstructure:"ENABLE_ELASTIC"`
	File          FileLoggerConfig `mapstructure:"FILE"`
	Elastic       ElasticConfig    `mapstructure:"ELASTIC"`
}

type FileLoggerConfig struct {
	Path string `mapstructure:"PATH"`
}

type ElasticConfig struct {
	URLs  []string `mapstructure:"URLS"`
	Index string   `mapstructure:"INDEX"`
}

// LoadConfig membaca konfigurasi dari environment variables.
func LoadConfig() (cfg Config, err error) {
	// Menetapkan nilai default
	viper.SetDefault("SERVER_LISTEN_ADDRESS", ":8080")
	viper.SetDefault("SERVER_BACKEND_URL", "http://localhost:3000")
	viper.SetDefault("UPLOADER_TYPE", "local")
	viper.SetDefault("UPLOADER_LOCAL_PATH", "/tmp/uploads")
	viper.SetDefault("LOGGER_FILE_PATH", "/tmp/interceptor.log")
	viper.SetDefault("LOGGER_ELASTIC_INDEX", "coraza-interceptor")

	// Mengaktifkan pembacaan dari environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Unmarshal konfigurasi ke struct
	err = viper.Unmarshal(&cfg)
	return
}
