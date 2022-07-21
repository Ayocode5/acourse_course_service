package contracts

import "acourse-course-service/pkg/models"

type CourseService interface {
	Fetch() ([]models.Course, error)
	FetchById(id string) (models.Course, error)
	Create() (models.Course, error)
}

type CourseDatabaseRepository interface {
	Fetch() (res []models.Course, err error)
	FetchById(id string) (res models.Course, err error)
	Create() (res models.Course, err error)
}
