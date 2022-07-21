package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type DBConnectionContract interface {
	GetDSN() string
	GetConnection() *mongo.Database
}

type Database struct {
	DbName       string
	DbHost       string
	DbPort       string
	DbCollection string
	Collection   *mongo.Collection
	Connection   DBConnectionContract
	connection   *mongo.Database
}

func (this *Database) Prepare() *Database {

	if this.connection == nil {

		clientOptions := options.Client().ApplyURI(this.GetDSN())

		client, err := mongo.NewClient(clientOptions)

		if err != nil {
			log.Fatal(err)
		}

		err = client.Connect(context.Background())

		if err != nil {
			log.Fatal(err)
		}

		this.connection = client.Database(this.DbName)

		log.Println("Connected to the database: MongoDB")
	} else {
		log.Println("Already Connected to the database: MongoDB")
	}

	return this
}

func (this *Database) GetConnection() *mongo.Database {
	return this.connection
}

func (this *Database) OpenCollection() *mongo.Collection {
	return this.connection.Collection(this.DbCollection)
}

func (this *Database) GetDSN() string {
	return fmt.Sprintf("mongodb://%s:%s/?compressors=disabled&gssapiServiceName=mongodb", this.DbHost, this.DbPort)
}
