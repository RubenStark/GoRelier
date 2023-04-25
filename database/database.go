package db

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var DB *gorm.DB

// ConnectDB is a function that connects to the database
func DBConnection() {
	var error error

	host := "localhost"
	port := "5432"
	user := "postgres"
	dbname := "postgres"
	password := "bloodboy123"

	DB, error = gorm.Open("postgres", "host="+host+" port="+port+" user="+user+" dbname="+dbname+" password="+password+" sslmode=disable")
	if error != nil {
		log.Fatal(error)
	} else {
		log.Println("Database connection successful")
	}
}
