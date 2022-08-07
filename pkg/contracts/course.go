package contracts

import (
	"acourse-course-service/pkg/http/requests"
	"acourse-course-service/pkg/models"
	"context"
)

type CourseService interface {
	Fetch(ctx context.Context) ([]models.Course, error)
	FetchById(ctx context.Context, id string) (models.Course, error)
	Create(ctx context.Context, data requests.CreateCourseRequest) (interface{}, error)
	Update(ctx context.Context, data requests.UpdateCourseRequest, course_id string) (res bool, err error)
}

type CourseDatabaseRepository interface {
	Fetch(ctx context.Context) (res []models.Course, err error)
	FetchById(ctx context.Context, id string) (res models.Course, err error)
	Create(ctx context.Context, data models.Course) (course_id string, err error)
	Update(ctx context.Context, data models.Course, course_id string) (res bool, err error)
}
