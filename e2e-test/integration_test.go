package e2e_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/gocql/gocql"
)

func waitForService(url string, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("service not ready after %d attempts", maxRetries)
}

func waitForServices(services map[string]string, maxRetries int) error {
	for service, url := range services {
		fmt.Printf("Waiting for %s...\n", service)
		if err := waitForService(url, maxRetries); err != nil {
			return fmt.Errorf("service %s not ready: %w", service, err)
		}
	}
	return nil
}

func checkKafkaLiveness(maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		cmd := exec.Command("docker", "exec", "kafka",
			"kafka-topics", "--create",
			"--if-not-exists",
			"--bootstrap-server", "kafka:9092",
			"--replication-factor", "1",
			"--partitions", "1",
			"--topic", "test-liveness")

		if err := cmd.Run(); err == nil {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("kafka not ready after %d attempts", maxRetries)
}

func TestEndToEndLogFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start the infrastructure using docker compose
	fmt.Println("Starting infrastructure services...")
	cmd := exec.Command("docker", "compose", "up", "-d", "kafka", "zookeeper", "clickhouse", "cassandra", "cockroachdb")
	cmd.Dir = ".." // Run from parent directory where docker-compose.yml is
	if err := cmd.Run(); err != nil {
		t.Fatal("Failed to start infrastructure:", err)
	}
	defer exec.Command("docker", "compose", "down").Run()

	fmt.Println("Checking Kafka connectivity...")
	if err := checkKafkaLiveness(40); err != nil {
		t.Fatal("Kafka not ready:", err)
	}

	// Wait for infrastructure services to be ready
	fmt.Println("Waiting for infrastructure services...")
	infraServices := map[string]string{
		"ClickHouse":  "http://localhost:8123/ping",
		"CockroachDB": "http://localhost:8082/health",
	}

	if err := waitForServices(infraServices, 40); err != nil {
		t.Fatal("Infrastructure services not ready:", err)
	}

	// Start application services
	services := []string{"log-drain", "cassandra-writer", "clickhouse-writer", "log-generator"}
	for _, service := range services {
		fmt.Printf("Starting %s...\n", service)
		cmd := exec.Command("docker", "compose", "up", "-d", service)
		cmd.Dir = ".."
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to start %s: %v", service, err)
		}
		time.Sleep(10 * time.Second) // Give each service more time to initialize
	}

	// Wait for log generator to be ready
	if err := waitForService("http://localhost:8084/health", 12); err != nil {
		t.Fatal("Log generator service not ready:", err)
	}

	// Trigger log generation
	fmt.Println("Triggering log generation...")
	resp, err := http.Post("http://localhost:8084/send-now", "", nil)
	if err != nil {
		t.Fatal("Failed to trigger log generation:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("Expected status OK, got %v with body: %s", resp.StatusCode, string(body))
	}

	// Wait for logs to propagate through the system
	time.Sleep(30 * time.Second)

	// Check data in ClickHouse
	fmt.Println("Verifying data in ClickHouse...")
	cmd = exec.Command("curl", "-s", "http://localhost:8123/", "--data-binary",
		"SELECT COUNT(*) as count, name FROM events GROUP BY name FORMAT JSON")
	cmd.Dir = ".."
	output, err := cmd.Output()
	if err != nil {
		t.Fatal("Failed to query ClickHouse:", err)
	}

	var clickhouseResult struct {
		Data []struct {
			Count int    `json:"count"`
			Name  string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(output, &clickhouseResult); err != nil {
		t.Fatal("Failed to parse ClickHouse response:", err)
	}
	if len(clickhouseResult.Data) == 0 {
		t.Error("No data found in ClickHouse")
	}

	// Check data in Cassandra
	fmt.Println("Verifying data in Cassandra...")
	cluster := gocql.NewCluster("localhost")
	cluster.Port = 9042
	cluster.Keyspace = "log_analyzer"
	cluster.Consistency = gocql.One
	cluster.ConnectTimeout = time.Second * 10

	session, err := cluster.CreateSession()
	if err != nil {
		t.Fatal("Failed to connect to Cassandra:", err)
	}
	defer session.Close()

	var count int
	if err := session.Query("SELECT COUNT(*) FROM events").Scan(&count); err != nil {
		t.Fatal("Failed to query Cassandra:", err)
	}
	if count == 0 {
		t.Error("No data found in Cassandra")
	}

	fmt.Printf("Test completed. Found %d events in Cassandra and %d distinct event types in ClickHouse\n",
		count, len(clickhouseResult.Data))
}
