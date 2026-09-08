package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds edge-daemon runtime settings loaded from the environment.
type Config struct {
	WellID       string        `env:"WELL_ID" envDefault:"WELL-DEMO-01"`
	SerialPort   string        `env:"SERIAL_PORT,required"`
	SerialBaud   int           `env:"SERIAL_BAUD" envDefault:"9600"`
	SQLitePath   string        `env:"SQLITE_PATH" envDefault:"./edge.db"`
	MQTTBroker   string        `env:"MQTT_BROKER" envDefault:"tcp://localhost:1883"`
	MQTTClientID string        `env:"MQTT_CLIENT_ID" envDefault:"edge-daemon"`
	MQTTQoS      byte          `env:"MQTT_QOS" envDefault:"1"`
	SyncInterval time.Duration `env:"SYNC_INTERVAL" envDefault:"1s"`
	SyncBatch    int           `env:"SYNC_BATCH" envDefault:"50"`
}

// Load parses environment variables into Config.
func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
