package common

var CockRoachdbAdress = "postgresql://username:password@hostname:26257/dbname?sslmode=require"

var RESTAPIPort = ":8081"

var RESTAPIBaseURL = "http://rest-api" + RESTAPIPort

var GUIBaseURL = ":8083"

var KafkaBrokers = []string{"kafka1:9092", "kafka2:9092", "kafka3:9092"}
