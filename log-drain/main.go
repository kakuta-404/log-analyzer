package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kakuta-404/log-analyzer/common"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

var kafkaWriter *kafka.Writer

func GenerateTopic(brokerAddr string, topic string, partitions int, replication int) {
	conn, err := kafka.Dial("tcp", brokerAddr)
	if err != nil {
		log.Printf("cannot dial kafka broker", err)
	}
	defer conn.Close()

	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: replication,
	})

	if err != nil {
		if !isTopicexcist(err) {
			log.Printf("topic exist", topic, err)
		}

	} else {
		log.Printf("topic created sucessfully", topic)
	}

}
func isTopicexcist(err error) bool {
	return err != nil && (err.Error() == "kafka: topic already exists" || err.Error() == "Topic with this name already exists")
}

func KafkaStarter() {
	brokerAddr := "kafka:9092"
	topic := "logs"
	GenerateTopic(brokerAddr, topic, 1, 1)
	kafkaWriter = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{brokerAddr},
		Topic:        topic,
		BatchSize:    1,
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
	log.Printf("Waiting 15 seconds to ensure Kafka is up")
	time.Sleep(15 * time.Second)

	GetInfo()
	KafkaStarter()
	log.Println("lionel messi")
	r := gin.Default()

	r.POST("/logs", func(c *gin.Context) {
		log.Println(" log called")
		var sub common.Submission

		if err := c.ShouldBindJSON(&sub); err != nil {
			log.Println("error in bnding json")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Println("Parsed submission succefully")

		// Call SendToKafka and handle the result
		if err := SendToKafka(&sub); err != nil {
			log.Printf("Error sending to Kafka: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to send log to Kafka",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Log submitted successfully"})
	})

	r.Run(":8080")
}

var PorjectsInfo map[string]string
var database sql.DB

func GetInfo() error {
	PorjectsInfo := make(map[string]string)
	database, err := sql.Open("postgres", "postgresql://root@cockroachdb:26257/defaultdb?sslmode=disable")
	if err != nil {
		log.Printf("could not connected to database", err)
		return err
	}
	log.Printf("connected to database")
	defer database.Close()

	rows, err2 := database.Query("SELECT * FROM projects")

	if err2 != nil {
		log.Printf("could not select projectID", err2)
		return err2
	}

	time.Sleep(20 * time.Second)

	log.Printf("selected")

	defer rows.Close()
	if rows == nil {
		log.Printf("rows are nil")
	}
	for rows.Next() {
		var project_id, api_key string
		if err3 := rows.Scan(&project_id, &api_key); err3 != nil {
			return err3
		}
		log.Printf("rows", project_id, api_key)
		PorjectsInfo[project_id] = api_key
	}
	return nil
}

func Authentication(key string, api_key string) (bool, error) {
	value, exist := PorjectsInfo[key]
	if !exist {
		var project_id, api_key string
		query := `SELECT projectid, name FROM projects WHERE project_id = $1 LIMIT 1`
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
