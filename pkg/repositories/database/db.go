package repositories

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/models"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ctx = context.Background()

type DatabaseRepository struct {
	Connection *mongo.Database
	Collection *mongo.Collection
}

func ConstructDBRepository(conn *mongo.Database, coll *mongo.Collection) contracts.CourseDatabaseRepository {

	return &DatabaseRepository{
		Connection: conn,
		Collection: coll,
	}
}

func (d DatabaseRepository) Fetch() (res []models.Course, err error) {

	//Fetch Connection
	records, err := d.Collection.Find(ctx, bson.M{})

	//Close Cursor
	defer func(records *mongo.Cursor, ctx context.Context) {
		err := records.Close(ctx)
		if err != nil {
			panic(err)
		}
	}(records, ctx)

	if err != nil {
		return nil, err
	}

	results := make([]models.Course, 0)

	//Append Each Record to results
	for records.Next(ctx) {
		var row models.Course

		err := records.Decode(&row)
		if err != nil {
			return nil, err
		}

		results = append(results, row)
	}

	return results, nil
}

func (d DatabaseRepository) FetchById(id string) (res models.Course, err error) {

	var result models.Course
	err = d.Collection.FindOne(ctx, bson.D{{"name", id}}).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (d DatabaseRepository) Create() (res models.Course, err error) {
	//TODO implement me
	panic("implement me")
}
