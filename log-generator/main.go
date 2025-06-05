package main

import (
	// "net/http"
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"time"
	"github.com/kakuto-404/log-analyzer/common"
	"github.com/lib/pq"
	// "github.com/gin-gonic/gin"
)

type Submission struct {
	ProjectID string            `json:"project_id"`
	APIKey    string            `json:"api_key"`
	Name      string            `json:"name"`
	Timestamp int32             `json:"timestamp"`
	PayLoad   map[string]string `json:"payload"`
}

func MakeEvent() Submission {
	var newEvent Submission
	newEvent.ProjectID = strconv.Itoa(RandomiseInteger())
	newEvent.APIKey = "lionel"
	newEvent.Name = RandomiseString()
	newEvent.Timestamp = int32(time.Now().UnixNano())
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
func main() {

	count := 0
	for count < 20 {
		newEvent := MakeEvent()
		fmt.Printf("Generated Event: %+v\n", newEvent)
		count++
	}

	// r := gin.Default()

	// r.POST("/logs", func(c *gin.Context) {
	// 	var payload LogPayload
	// 	if err := c.BindJSON(&payload); err != nil {
	// 		c.JSON(400, gin.H{"error": err.Error()})
	// 		return
	// 	}
	// 	// TODO: Validate API key and send to Kafka
	// 	c.JSON(200, gin.H{"status": "received"})
	// })

	// r.Run(":8080")
}
