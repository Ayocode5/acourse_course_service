package contracts

type CourseStorageRepository interface {
	Put()
	Fetch()
	Delete()
}
