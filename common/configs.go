package common

// a golobal adress var for connecting to cockroachdb for handel possible future errors

var CockRoachdbAdress = "postgresql://username:password@hostname:26257/dbname?sslmode=require"

var RESTAPIBaseURL = "http://localhost" + RESTAPIPort // ✅ For clients

var RESTAPIPort = ":8081" // ✅ For servers

var GUIBaseURL = ":8082"
