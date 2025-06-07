filter by event
prune data in clickhouse writer
every logger in clickhouse-writer/writer/writer.go is info or higher level this is not a good practice but the configs did not load properly during the integration test -> fix logging level and configs of clickhouse-writer