package config

import "time"

type Config struct {
	Server struct {
		Port string
	}
	Generator struct {
		Interval    time.Duration
		LogDrainURL string
	}
	CockroachDB struct {
		DSN string
	}
}

func Load() (*Config, error) {
	cfg := &Config{}
	cfg.Server.Port = ":8084"
	cfg.Generator.Interval = time.Second
	cfg.Generator.LogDrainURL = "http://log-drain:8080/logs"
	cfg.CockroachDB.DSN = "postgresql://root@cockroachdb:26257/defaultdb?sslmode=disable"
	return cfg, nil
}
