# System Architecture and Requirements

## Overview
This project implements a scalable log analysis system using Cassandra, Kafka, ClickHouse, and CockroachDB. The system provides a simple user interface with a highly scalable architecture for log aggregation and analysis.

## System Requirements

### User Management
- Users have username and password
- Users can access multiple projects

### Project Configuration
- Projects have name, searchable keys, API key, and TTL
- Log entries are automatically deleted after TTL
- Projects define searchable fields at creation (immutable)

### Log Ingestion API
- Requires project_id and API key authentication
- Payload includes:
  - Event name (for grouping)
  - Timestamp
  - Key-value pairs (mix of searchable and non-searchable)
  - All values stored as strings

### Log Viewing Interface
- Search capabilities:
  - Filter by searchable keys
  - View all events of a name without filters
- Event summary table showing:
  - Event name
  - Last occurrence
  - Total count within filter
- Detailed event view:
  - All key-value pairs
  - Occurrence time
  - Registration time
  - Navigation between events

### Performance Requirements
- Support for millions of events
- Short search response for searchable keys
- Appropriate consistency levels for different data

### Technical Requirements
- Backend in GoLang
- Use of all 4 databases: Cassandra, Kafka, ClickHouse, CockroachDB
- Fault-tolerant design
- Replication factor of 3 for data stores (except ClickHouse data for deployment simplicity)

## Deliverables
1. Design Documentation
   - System component diagrams
   - Rationale for architectural choices (pros and cons)
   - Data structure definitions (table schemas)
2. System Performance Evaluation
   - Tests demonstrating performance meets requirements
3. Implementation
4. Test Log Generator
   - Tool for generating high volumes of test logs for different projects
5. Demo Video (max 15 minutes)
   - Demonstration of system functionality and fault tolerance

## Simplifications
- ClickHouse deployed as a single node.
- Web server deployed as a stateless, single node.
- No need to test adding new nodes to the system.
- Data transfer from Kafka to ClickHouse should be implemented in Go.

```
╔════════════════════════════════════════════════════════════════════════╗
║                         SYSTEM ARCHITECTURE                            ║
╚════════════════════════════════════════════════════════════════════════╝

                      +-------------+                                    
                      |  Generate   |                                    
                      | (Test Logs) |                                    
                      +------+------+                                    
                             |                                           
                       +-----v-----+                                     
                       |    API    |                                     
                       +-----+-----+                                     
                             |                                           
                       +-----v-----+                                     
                       | Log-Drain |------------------------------+      
                       +-----+-----+                              |      
                             |                                    |      
                       +-----v-----+                              |      
                       |   Kafka   |                              |      
                       +-----+-----+                              |      
                             |                                    |      
               +-------------+-------------+                      |      
               |                           |                      |      
     +---------v---------+       +---------v---------+            |      
     | ClickHouse Writer |       | Cassandra Writer  |            |      
     +---------+---------+       +---------+---------+            |      
               |                           |                      |      
   +-----------v----------+     +----------v-----------+   +------v------+
   |    ClickHouseDB      |     |     CassandraDB      |   | CockroachDB |
   |(per-project indexing)|     |  (retrieve by name)  |   |    (Auth)   |
   +--------+-------------+     +--------------+-------+   +------^------+
             \                                /                   |      
              \                              /                    |      
               \                            /                     |      
                +------------+------------+                       |      
                             |                                    |      
                       +-----v-----+                              |      
                       |  Rest API |------------------------------+      
                       +-----+-----+                                     
                             |                                           
                          +-----+                                        
                          | GUI |                                        
                          +--+--+                                       
```

## Design Rationale

### Message Queue
A message queue is employed to decouple log ingestion from processing. It provides:
- Asynchronous processing of high-volume log events.
- Load buffering to prevent system overload.
- Increased fault tolerance by queuing logs in case downstream services are temporarily unavailable.

### ClickHouse
ClickHouse is chosen for its support for dynamic searchable keys. Its columnar storage and fast analytical query capabilities allow:
- Efficient querying over large datasets.
- Flexible schema management to adapt to evolving log formats.

### Cassandra
Cassandra is used to store raw log data and deliver the most up-to-date results. It is ideal for:
- Handling high write throughput.
- Fast searches on predefined keys (name in our case).
- Scalability across distributed environments, ensuring data remains current.

### CockroachDB
CockroachDB is adopted for managing authentication and metadata because:
- It uses familiar SQL with robust consistency guarantees.
- Its transactional support simplifies managing critical user data.
- It provides reliable fault tolerance for metadata operations.
