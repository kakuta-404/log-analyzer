filter by event
prune data in clickhouse writer
websocket instead of http connection
kafka async
every logger in clickhouse-writer/writer/writer.go is info or higher level this is not a good practice but the configs did not load properly during the integration test -> fix logging level and configs of clickhouse-writer
increase 'replication_factor' of cassandra
add ttl as project configs
add table name and keyspace names to config



 it not to do : 
 // another things that might come handy 
	/*
	indexSize := len(searchableKeys)
	if indexSize == 0 {
		indexSize = 100 // Default fallback if no keys
	}
	alterQuery := fmt.Sprintf(`
        ALTER TABLE %s
        ADD INDEX IF NOT EXISTS searchable_keys_index (mapKeys(log_data)) TYPE set(%d) GRANULARITY 1
    `, projectID, indexSize)

	if err := conn.Exec(context.Background(), alterQuery); err != nil {
		return fmt.Errorf("failed to add index to table %s: %w", projectID, err)
	}

	bloomQuery := fmt.Sprintf(`
        ALTER TABLE %s
        ADD INDEX IF NOT EXISTS searchable_keys_bloom (mapKeys(log_data)) TYPE bloom_filter(0.01) GRANULARITY 1
    `, projectID)

	if err := conn.Exec(context.Background(), bloomQuery); err != nil {
		return fmt.Errorf("failed to add bloom filter index to table %s: %w", projectID, err)
	}
	*/