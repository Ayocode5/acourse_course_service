package migration

import (
	"acourse-course-service/pkg/database"
)

type Migration struct {
	DB *database.Database
}

func ConstructMigration(db *database.Database) *Migration {
	return &Migration{DB: db}
}
