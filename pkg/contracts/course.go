package contracts

import (
	"acourse-course-service/pkg/http/middleware"
	"acourse-course-service/pkg/http/requests"
	"acourse-course-service/pkg/http/response"
	"acourse-course-service/pkg/models"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CourseService interface {
	CourseResourcePolicy
	Fetch(ctx context.Context, excludedFields []string, pagination models.Pagination) ([]models.Course, error)
	FetchById(ctx context.Context, id string, excludeFields []string) (models.Course, error)
	Create(ctx context.Context, data requests.CreateCourseRequest) (interface{}, error)
	Update(ctx context.Context, data requests.UpdateCourseRequest, course_id string) (*response.HttpResponse, error)
	DeleteMaterials(ctx context.Context, course_id string, data requests.DeleteMaterialsRequest) (*response.HttpResponse, error)
	DeleteCourse(ctx context.Context, course_id string) (*response.HttpResponse, error)
}

type CourseDatabaseRepository interface {
	Fetch(ctx context.Context, excludeFields []string, limit int64, skip int64) (res []models.Course, err error)
	FetchById(ctx context.Context, id string, excludeFields []string) (res models.Course, err error)
	FetchByUserId(ctx context.Context, user_id int64, excludeFields []string) (res *models.Course, err error)
	Create(ctx context.Context, data *models.Course) (course_id primitive.ObjectID, err error)
	Update(ctx context.Context, data models.Course, course_id string) (res bool, err error)
	DeleteCourse(ctx context.Context, course_id string) (res bool, err error)
	DeleteMaterials(ctx context.Context, course_id string, material_id []string) (res interface{}, err error)
	GenerateModelID() primitive.ObjectID
}

type CourseResourcePolicy interface {
	AuthorizeResourceByUserId(model *models.Course, authorization middleware.Authorization) (bool, error)
}
