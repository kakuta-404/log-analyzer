package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	"github.com/lib/pq"
)

func MakeSubmission() common.Submission {
	var newEvent common.Submission
	newEvent.ProjectID = strconv.Itoa(RandomiseInteger())
	newEvent.APIKey = "lionel"
	newEvent.Name = RandomiseString()
	newEvent.Timestamp = time.Now()
	newEvent.PayLoad = makePayload()
	return newEvent
}

func makePayload() map[string]string {
	payload := make(map[string]string)
	length := RandomiseInteger()
	for i := 0; i < length; i++ {
		payload[RandomiseString()] = RandomiseString()
	}
	return payload
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomiseInteger() int {
	return rand.Intn(61)
}

func RandomiseString() string {
	rand.Seed(time.Now().UnixNano())
	randomInt := RandomiseInteger()
	result := make([]byte, randomInt)
	for i := range result {
		result[i] = letters[RandomiseInteger()]
	}
	return string(result)
}

var Names []ProjectGe

// hardcoding for now

type ProjectGe struct {
	ProjectId     string   `json:"project_id"`
	APIKey        string   `json:"apikey"`
	SearchAbleKey []string `json:"search_able_key"`
}

// also get infoes for projects name and IDs
func ConnectToCockroachDB() error {
	connStr := "postgresql://root@localhost:26257/defaultdb?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("could not connect to cockraochDb", err)
		return err
	}
	log.Println("connected to the cockroachdb")
	defer db.Close()
	_, err1 := db.Exec(`
    CREATE TABLE IF NOT EXISTS projects (
        project_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        apikey STRING NOT NULL,
        searchable_keys STRING[]
    );
`)
	if err1 != nil  {
		log.Printf("could not created the desiarble table", err1)
	}

	log.Printf("sccessfuly created the table")

	_, err1 = db.Exec(`
    INSERT INTO projects (apikey, searchable_keys)
    VALUES ($1, $2)
`, "test-api-key", pq.StringArray{"lionl", "lionelmessi", "timestamp"})
	if err1 != nil {
		log.Printf("could not insert sample",err1)
	}
	rows, err := db.Query("SELECT project_id, apikey, searchable_keys FROM projects")
	if err != nil {
		return err
	}
	for rows.Next() {
		var p ProjectGe
		var keys pq.StringArray

		if err := rows.Scan(&p.ProjectId, &p.APIKey, &keys); err != nil {
			return err
		}

		p.SearchAbleKey = []string(keys)
		Names = append(Names, p)
		log.Printf("test ",p.SearchAbleKey,p.APIKey,p.ProjectId)
	}
	return nil

}

func sendLogs() {

	sub := MakeSubmission()
	body, _ := json.Marshal(sub)

	resp, err := http.Post("http://log-drain:8080/logs", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return
	}
	log.Println("log sent successfully ++++++ ___ -dlfk")
	resp.Body.Close()

}
func main() {
	ConnectToCockroachDB()
	go func() {
		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			log.Println("ticker triggered, sending log .....")
			sendLogs()
		}
	}()

	r := gin.Default()

	r.POST("/send-now", func(c *gin.Context) {
		sendLogs()
		c.JSON(http.StatusOK, gin.H{"status": "log sent manually"})
	})
	log.Println("Starting HTTP server on :8080")
	r.Run()

}
