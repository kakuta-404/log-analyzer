package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

var kafkaWriter *kafka.Writer

func init() {
	kafkaWriter = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{"kafka:9092"},
		Topic:        "logs",
		BatchSize:    500,
		BatchTimeout: 200 * time.Millisecond,
		Balancer:     &kafka.LeastBytes{},
	})
}

func SendToKafka(sub *common.Submission) error {
	// but it must check the bool value for authentication
	_, event := ConvertSubmissionTOEvent(sub)

	valueBytes, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}

	err = kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   nil,
		Value: valueBytes,
	})

	if err != nil {
		return err
	}

	return nil
}

func main() {
	r := gin.Default()

	r.POST("/logs", func(c *gin.Context) {
		var sub common.Submission

		if err := c.ShouldBindJSON(&sub); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		SendToKafka(&sub)

	})

	r.Run(":8080")
}

var PorjectsInfo map[string]string
var database sql.DB

func GetInfo() error {
	PorjectsInfo := make(map[string]string)
	database, err := sql.Open("postgres", common.CockRoachdbAdress)
	if err != nil {
		return err
	}
	defer database.Close()

	rows, err2 := database.Query("SELECT ProjectID, APIKEY FROM projects")
	if err2 != nil {
		return err2
	}
	defer rows.Close()
	for rows.Next() {
		var project_id, api_key string
		if err3 := rows.Scan(&project_id, &api_key); err3 != nil {
			return err3
		}
		PorjectsInfo[project_id] = api_key
	}
	return nil
}

func Authentication(key string, api_key string) (bool, error) {
	value, exist := PorjectsInfo[key]
	if exist == false {
		var project_id, api_key string
		query := `SELECT projectid, name FROM projects WHERE projectid = $1 LIMIT 1`
		err := database.QueryRow(query, value).Scan(&project_id, &api_key)

		if err == sql.ErrNoRows {
			return false, err
		}
		if err != nil {
			return false, err
		}

		PorjectsInfo[key] = value
	}
	if value == api_key {
		return true, nil
	} else {
		return false, nil
	}

}

func ConvertSubmissionTOEvent(sub *common.Submission) (bool, common.Event) {
	// check authentication
	var event common.Event
	// event.Name = "lionel"
	// if auth, _ :=Authentication(sub.ProjectID, sub.APIKey); auth  == false{
	// 	return false,event
	// }
	event.Log = sub.PayLoad
	event.ProjectID = sub.ProjectID
	event.Name = sub.Name
	event.EventTimestamp = sub.Timestamp
	return true, event
}
