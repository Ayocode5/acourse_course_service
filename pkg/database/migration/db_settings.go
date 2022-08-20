package migration

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func (m Migration) MigrateSettings() {
	m.CreateIndexes()
	log.Println("Migrates Settings Success")
}

func (m Migration) CreateIndexes() {
	//_, err := m.DB.GetCollection().Indexes().DropOne(context.Background(), "user_id_1")
	//if err != nil {
	//	panic(err)
	//}

	_, err := m.DB.GetCollection().Indexes().CreateOne(context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(false),
		})
	if err != nil {
		panic(err)
	}
}
