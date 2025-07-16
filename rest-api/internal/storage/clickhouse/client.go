package clickhouse

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2"
)

var Conn clickhouse.Conn

func Init(host string) error {
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
		return err
	}

	return Conn.Ping(context.Background())
}
