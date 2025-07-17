package e2e_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"testing"
	"time"
)

func verifyAPIResponse(t *testing.T, url string) {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal("Failed to call API endpoint:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Failed to read response body:", err)
	}
	t.Logf("Response from %s: %s", url, string(body))

	// Re-create reader for JSON decoding
	result := map[string]interface{}{}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		t.Fatal("Failed to decode JSON response:", err)
	}

	groups, ok := result["groups"].([]interface{})
	if !ok || len(groups) == 0 {
		t.Fatal("Expected at least one group in response")
	}
}

func TestEndToEnd(t *testing.T) {
	cmd := exec.Command("docker", "compose", "up", "-d", "kafka", "clickhouse", "cassandra", "cockroachdb")
	cmd.Dir = ".."
	if err := cmd.Run(); err != nil {
		t.Fatal("Failed to start services:", err)
	}
	defer exec.Command("docker", "compose", "down").Run()

	// Wait for infrastructures to be available
	time.Sleep(60 * time.Second)

	logDrainCmd := exec.Command("docker", "compose", "up", "-d", "log-drain")
	logDrainCmd.Dir = ".."
	if err := logDrainCmd.Start(); err != nil {
		t.Fatal("Failed to start log-drain service:", err)
	}
	time.Sleep(15 * time.Second)

	appCmd := exec.Command("docker", "compose", "up", "-d", "log-generator", "clickhouse-writer", "cassandra-writer", "rest-api")
	appCmd.Dir = ".."
	if err := appCmd.Start(); err != nil {
		t.Fatal("Failed to start app services:", err)
	}

	time.Sleep(30 * time.Second)

	verifyAPIResponse(t, "http://localhost:8081/projects/test-project/events")
	verifyAPIResponse(t, "http://localhost:8081/projects/test-project/search?registered=true")
}
