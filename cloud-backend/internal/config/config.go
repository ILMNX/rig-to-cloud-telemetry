package config

import (
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config holds cloud ingest/API runtime settings.
type Config struct {
	DatabaseURL string   `env:"DATABASE_URL" envDefault:"postgres://telemetry:telemetry@localhost:5432/telemetry?sslmode=disable"`
	MQTTBroker  string   `env:"MQTT_BROKER" envDefault:"tcp://localhost:1883"`
	MQTTClientID string  `env:"MQTT_CLIENT_ID" envDefault:"cloud-ingest"`
	MQTTQoS     byte     `env:"MQTT_QOS" envDefault:"1"`
	HTTPAddr    string   `env:"HTTP_ADDR" envDefault:":8080"`
	CORSOrigins []string `env:"CORS_ORIGINS" envDefault:"http://localhost:5173" envSeparator:","`
}

// Load parses environment variables into Config.
func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	for i, o := range cfg.CORSOrigins {
		cfg.CORSOrigins[i] = strings.TrimSpace(o)
	}
	return cfg, nil
}
