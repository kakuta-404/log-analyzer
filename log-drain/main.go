package main

import (
	"encoding/json"
	"log"
	"github.com/kakuta-404/log-analyzer/common"
	_ "github.com/lib/pq"
	"database/sql"

)

type LogProcessor struct {
	// TODO: Add Kafka consumer configuration
}

func main() {
	log.Println("Starting Log Drain Service...")
	// TODO: Initialize Kafka consumer and start processing
}

func (p *LogProcessor) processLog(data []byte) error {
	var event common.Event
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}
	// TODO: Process event and route to appropriate writer
	return nil
}

func ConnectToLogGenerator() {
	

}

var PorjectsInfo map[string]string
var database sql.DB

func GetInfo() error  {
	PorjectsInfo := make(map[string]string)
	database ,err  := sql.Open("postgres",common.CockRoachdbAdress)
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
		if err3 := rows.Scan(&project_id,&api_key); err3 != nil {
			return err3
		}
		PorjectsInfo[project_id] = api_key
	}
	return nil
}

func Authentication(key string, api_key string) (bool, error) {
	value, exist := PorjectsInfo[key]
	if exist == false {
		var project_id,api_key string
		query := `SELECT projectid, name FROM projects WHERE projectid = $1 LIMIT 1`
   		 err := database.QueryRow(query, value).Scan(&project_id, &api_key)

    	if err == sql.ErrNoRows {
        	return false,err
    	} 
		if err != nil  {
			return false, err
		}

		PorjectsInfo[key] = value
	}
	if value == api_key {
		return true,nil
	} else {
		return false,nil
	}
	
}

func SendToKafka() {

}