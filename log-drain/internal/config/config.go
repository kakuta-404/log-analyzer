package config

type Config struct {
	Server struct {
		Port string
	}
	Kafka struct {
		Brokers []string
		Topic   string
	}
	CockroachDB struct {
		DSN string
	}
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Default values
	cfg.Server.Port = ":8080"
	cfg.Kafka.Brokers = []string{"kafka:9092"}
	cfg.Kafka.Topic = "logs"
	cfg.CockroachDB.DSN = "postgresql://root@cockroachdb:26257/defaultdb?sslmode=disable"

	return cfg, nil
}
