package repositories

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/models"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

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

func (d DatabaseRepository) Fetch(ctx context.Context) (res []models.Course, err error) {

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

func (d DatabaseRepository) FetchById(ctx context.Context, id string) (res models.Course, err error) {

	var result models.Course
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return result, err
	}
	err = d.Collection.FindOne(ctx, bson.D{{"_id", objectID}}).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (d DatabaseRepository) Create(ctx context.Context, data models.Course) (string_id string, err error) {

	var course_id string

	//	Use Transaction
	err = d.Connection.Client().UseSession(ctx, func(sessionContext mongo.SessionContext) error {
		// Start Transaction
		err := sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		// Insert Data To the Database & abort if it fails
		insertedData, err := d.Collection.InsertOne(ctx, data)
		if err != nil {
			sessionContext.AbortTransaction(ctx)
			return err
		}

		course_id = insertedData.InsertedID.(primitive.ObjectID).Hex()

		// Commit Data if no error
		err = sessionContext.CommitTransaction(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return course_id, nil
}

func (d DatabaseRepository) Update(ctx context.Context, data models.Course, id string) (res bool, err error) {

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}
	filter := bson.D{{"_id", objectId}}
	_, err = d.Collection.UpdateOne(ctx, filter, bson.D{{"$set", data}})
	if err != nil {
		return false, err
	}
	return true, err
}
