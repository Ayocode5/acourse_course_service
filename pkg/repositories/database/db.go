package repositories

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/models"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (d DatabaseRepository) Fetch(ctx context.Context, excludeFields []string) (res []models.Course, err error) {

	//Exclude fields
	excluded := make(map[string]int)
	for _, field := range excludeFields {
		excluded[field] = 0
	}

	opts := options.Find().SetProjection(excluded)

	//Fetch Records
	records, err := d.Collection.Find(ctx, bson.M{}, opts)

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

func (d DatabaseRepository) FetchById(ctx context.Context, id string, excludeFields []string) (res models.Course, err error) {

	//Exclude fields
	excluded := make(map[string]int)
	for _, field := range excludeFields {
		excluded[field] = 0
	}

	opts := options.FindOne().SetProjection(excluded)

	var result models.Course
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return result, err
	}
	err = d.Collection.FindOne(ctx, bson.D{{"_id", objectID}}, opts).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (d DatabaseRepository) FetchByUserId(ctx context.Context, user_id int64) (res models.Course, err error) {

	var result models.Course
	err = d.Collection.FindOne(ctx, bson.D{{"user_id", user_id}}).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (d DatabaseRepository) Create(ctx context.Context, data models.Course) (string_id primitive.ObjectID, err error) {

	var course_id primitive.ObjectID

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

		course_id = insertedData.InsertedID.(primitive.ObjectID)

		// Commit Data if no error
		err = sessionContext.CommitTransaction(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return primitive.NilObjectID, err
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

func (d DatabaseRepository) DeleteCourse(ctx context.Context, course_id string) (res bool, err error) {
	objectID, err := primitive.ObjectIDFromHex(course_id)
	if err != nil {
		return false, err
	}
	_, err = d.Collection.DeleteOne(ctx, bson.D{{"_id", objectID}})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d DatabaseRepository) DeleteMaterials(ctx context.Context, course_id string, material_id []string) (res interface{}, err error) {

	var material_ids []primitive.ObjectID

	for _, m_id := range material_id {
		objectID, err := primitive.ObjectIDFromHex(m_id)
		if err != nil {
			return false, err
		}
		material_ids = append(material_ids, objectID)
	}

	pull := bson.D{{"$pull", bson.D{{"materials", bson.D{{"material_id", bson.D{{"$in", material_ids}}}}}}}}

	objectID, err := primitive.ObjectIDFromHex(course_id)
	if err != nil {
		return false, err
	}

	filter := bson.D{{"_id", objectID}}

	res, err = d.Collection.UpdateOne(ctx, filter, pull)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (d DatabaseRepository) GenerateModelID() primitive.ObjectID {
	return primitive.NewObjectID()
}
