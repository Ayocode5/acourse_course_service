package database

import (
	"acourse-course-service/pkg/contracts"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type Database struct {
	DbUsername   string
	DBPassword   string
	DbName       string
	DbHost       string
	DbPort       string
	DbCollection string
	collection   *mongo.Collection
	connection   *mongo.Database
}

func (db *Database) Prepare() contracts.MongoDBContract {

	if db.connection == nil {

		clientOptions := options.Client().ApplyURI(db.Dsn())

		client, err := mongo.NewClient(clientOptions)

		if err != nil {
			panic(err.Error())
		}

		err = client.Connect(context.Background())

		if err != nil {
			panic(err.Error())
		}

		log.Println("PINGING: MongoDB")
		err = client.Ping(context.Background(), nil)
		if err != nil {
			panic(err)
		}

		db.connection = client.Database(db.DbName)

		log.Println("Connected to the database: MongoDB")
	} else {
		log.Println("Already Connected to the database: MongoDB")
	}

	return db
}

func (db *Database) GetConnection() *mongo.Database {
	return db.connection
}

func (db *Database) GetCollection() *mongo.Collection {
	return db.connection.Collection(db.DbCollection)
}

func (db *Database) Dsn() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin", db.DbUsername, db.DBPassword, db.DbHost, db.DbPort, db.DbName)
}
