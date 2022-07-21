package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type courses struct {
	Name string `bson:"name"`
}

func AllCourse() []courses {

	records, err := collection.Find(ctx, bson.M{})
	defer func(records *mongo.Cursor, ctx context.Context) {
		err := records.Close(ctx)
		if err != nil {
			panic(err)
		}
	}(records, ctx)

	if err != nil {
		log.Fatalf(err.Error())
	}

	result := make([]courses, 0)

	for records.Next(ctx) {
		var row courses

		err := records.Decode(&row)
		if err != nil {
			log.Fatal(err.Error())
		}

		result = append(result, row)
	}

	return result
}

func FindCourse(id string) (*courses, error) {
	var result courses
	err := collection.FindOne(ctx, bson.D{{"name", id}}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
