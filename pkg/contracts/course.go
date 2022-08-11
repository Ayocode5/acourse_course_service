package contracts

import (
	"acourse-course-service/pkg/http/requests"
	"acourse-course-service/pkg/models"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CourseService interface {
	Fetch(ctx context.Context, excludedFields []string) ([]models.Course, error)
	FetchById(ctx context.Context, id string, excludeFields []string) (models.Course, error)
	Create(ctx context.Context, data requests.CreateCourseRequest) (interface{}, error)
	Update(ctx context.Context, data requests.UpdateCourseRequest, course_id string) (bool, error)
	DeleteMaterials(ctx context.Context, course_id string, material_id []string) (bool, error)
	DeleteCourse(ctx context.Context, course_id string) (bool, error)
}

type CourseDatabaseRepository interface {
	Fetch(ctx context.Context, excludeFields []string) (res []models.Course, err error)
	FetchById(ctx context.Context, id string, excludeFields []string) (res models.Course, err error)
	FetchByUserId(ctx context.Context, user_id int64) (res models.Course, err error)
	Create(ctx context.Context, data models.Course) (course_id primitive.ObjectID, err error)
	Update(ctx context.Context, data models.Course, course_id string) (res bool, err error)
	DeleteCourse(ctx context.Context, course_id string) (res bool, err error)
	DeleteMaterials(ctx context.Context, course_id string, material_id []string) (res interface{}, err error)
	GenerateModelID() primitive.ObjectID
}
