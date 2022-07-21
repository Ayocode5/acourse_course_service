package contracts

import "go.mongodb.org/mongo-driver/mongo"

type DBContract interface {
	Dsn() string
	GetConnection() *mongo.Database
}

type MongoDBContract interface {
	GetCollection() *mongo.Collection
	DBContract
}
