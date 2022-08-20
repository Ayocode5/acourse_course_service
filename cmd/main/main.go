package main

import (
	"acourse-course-service/pkg/database"
	migrations "acourse-course-service/pkg/database/migration"
	"acourse-course-service/pkg/http/controllers"
	dbrepo "acourse-course-service/pkg/repositories/database"
	s3repo "acourse-course-service/pkg/repositories/storage"
	"acourse-course-service/pkg/services"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
)

import "github.com/zhulik/go_mediainfo"

func main() {

	var ctx = context.Background()

	//Load .Env
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	//Create Gin Instance
	engine := gin.Default()

	//Setup Database
	db := database.Database{
		DbName:       os.Getenv("DB_NAME"),
		DbCollection: os.Getenv("DB_COLLECTION"),
		DbHost:       os.Getenv("DB_HOST"),
		DbPort:       os.Getenv("DB_PORT"),
		DbUsername:   os.Getenv("DB_USERNAME"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
	}
	//Open Connection
	mongodb := db.Prepare()

	//Migrations
	migration := migrations.ConstructMigration(&db)
	migration.MigrateSettings()

	//Setup MongoDB Repository
	dbRepository := dbrepo.ConstructDBRepository(mongodb.GetConnection(), mongodb.GetCollection())

	//Setup S3 Storage Repository
	s3StorageRepository := s3repo.ConstructS3Repository(
		os.Getenv("AWS_ACCESS_KEY_ID"),
		os.Getenv("AWS_SECRET_ACCESS_KEY"),
		os.Getenv("AWS_BUCKET_NAME"),
		os.Getenv("AWS_BUCKET_REGION"),
		[]string{"video/mp4", "video/x-matroska", "image/jpeg"},
		int64(100*1024*1024), //MAX Filesize 100mb
		3,
	)

	//Setup Storage CourseService
	storageService := services.ConstructStorageService(&s3StorageRepository)

	//Setup MediaInfo Service
	mediaInfoService := services.ConstructMediaInfoService(mediainfo.NewMediaInfo())

	//Setup Course Services
	courseService := services.ConstructCourseService(&dbRepository, &storageService, &mediaInfoService)

	//Setup Course Devlivery/Http Controller
	controllers.SetupCourseHandler(ctx, engine, courseService)

	//Running App With Desired Port
	if port := os.Getenv("APP_PORT"); port == "" {
		err := engine.Run(":8080")
		if err != nil {
			panic(err)
		}
	} else {
		err := engine.Run(":" + port)
		if err != nil {
			panic(err)
		}
	}

}
