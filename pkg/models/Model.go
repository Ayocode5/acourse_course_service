package models

import (
	"acourse-course-service/pkg/database"
	"context"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
)

var connection *mongo.Database
var collection *mongo.Collection
var ctx = context.Background()

func init() {

	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	//Setup Connection
	db := database.Database{
		DbName:       os.Getenv("DB_NAME"),
		DbCollection: os.Getenv("DB_COLLECTION"),
		DbHost:       os.Getenv("DB_HOST"),
		DbPort:       os.Getenv("DB_PORT"),
	}

	prepared := db.Prepare()
	connection = prepared.GetConnection()
	collection = prepared.OpenCollection()
}
