package services

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/models"
)

type CourseService struct {
	DBRepository contracts.CourseDatabaseRepository
}

func ConstructCourseService(dbRepository contracts.CourseDatabaseRepository) contracts.CourseService {
	return &CourseService{DBRepository: dbRepository}
}

func (c CourseService) Fetch() ([]models.Course, error) {
	return c.DBRepository.Fetch()
}

func (c CourseService) FetchById(id string) (models.Course, error) {
	return c.DBRepository.FetchById(id)
}

func (c CourseService) Create() (models.Course, error) {
	//TODO implement me
	panic("implement me")
}
