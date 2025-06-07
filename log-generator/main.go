package main

import (
	// "net/http"
	"database/sql"
	"math/rand"
	"strconv"
	"github.com/lib/pq"
	"github.com/kakuta-404/log-analyzer/common"
	"log"
	"net/url"
	"time"
	"github.com/gorilla/websocket"
	// "github.com/gin-gonic/gin"
)

func MakeSubmission() common.Submission {
	var newEvent common.Submission
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

	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:8080",
		Path:   "/ws",
	}

	log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()
	log.Println("Connected to log-drain")



	for {
		var sub common.Submission 
		sub = MakeSubmission()
		err = conn.WriteJSON(sub)
		if err != nil {
		log.Fatalf("Failed to send: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}