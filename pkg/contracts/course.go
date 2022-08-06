package contracts

import (
	"acourse-course-service/pkg/http/requests"
	"acourse-course-service/pkg/models"
)

type CourseService interface {
	Fetch() ([]models.Course, error)
	FetchById(id string) (models.Course, error)
	Create(data requests.CreateCourseRequest) (interface{}, error)
}

type CourseDatabaseRepository interface {
	Fetch() (res []models.Course, err error)
	FetchById(id string) (res models.Course, err error)
	Create(data models.Course) (course_id string, err error)
}
