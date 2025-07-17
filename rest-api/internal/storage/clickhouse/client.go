package clickhouse

import (
	"context"
	"log/slog"

	"github.com/ClickHouse/clickhouse-go/v2"
)

var Conn clickhouse.Conn

func Init(host string) error {
	slog.Info("Initializing ClickHouse connection", "host", host)
	var err error
	Conn, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{host},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
	})
	if err != nil {
		slog.Error("Error creating ClickHouse connection", "error", err)
		return err
	}

	if err := Conn.Ping(context.Background()); err != nil {
		slog.Error("Error pinging ClickHouse", "error", err)
		return err
	}

	slog.Info("Successfully connected to ClickHouse")
	return nil
}
