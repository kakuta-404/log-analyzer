package main

import (
	"net/http"
	"database/sql"
	"math/rand"
	"strconv"
	"github.com/lib/pq"
	"github.com/kakuta-404/log-analyzer/common"
	"time"
	"encoding/json"
	"bytes"
	"github.com/gin-gonic/gin"
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
	connStr := "postgresql://<user>:<password>@<host>:<port>/<database>?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()
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
    resp.Body.Close()


}
func main() {
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        for range ticker.C {
            sendLogs()
        }
    }()

    r := gin.Default()

    r.POST("/send-now", func(c *gin.Context) {
        sendLogs()
        c.JSON(http.StatusOK, gin.H{"status": "log sent manually"})
    })


}