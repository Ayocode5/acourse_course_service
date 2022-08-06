package main

import (
	"acourse-course-service/pkg/database"
	"acourse-course-service/pkg/http/controllers"
	dbrepo "acourse-course-service/pkg/repositories/database"
	s3repo "acourse-course-service/pkg/repositories/storage"
	"acourse-course-service/pkg/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
)

import "github.com/zhulik/go_mediainfo"

func main() {

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
	}
	//Open Connection
	mongodb := db.Prepare()

	//Setup MongoDB Repository
	dbRepository := dbrepo.ConstructDBRepository(mongodb.GetConnection(), mongodb.GetCollection())

	//Setup S3 Storage Repository
	s3StorageRepository := s3repo.ConstructS3Repository(
		"AKIA2UG6E7BNH4BCHPD3",
		"3otZOFOcuQBB2xLnDp/T7LSDFu2Ko/qJU4Lx8QPJ",
		"acourse-course-videos",
		"ap-southeast-1",
		[]string{"video/mp4", "video/x-matroska", "image/jpeg"},
		int64(100*1024*1024),
		3,
	)

	//Setup Storage CourseService
	storageService := services.ConstructStorageService(&s3StorageRepository)

	//Setup MediaInfo Service
	mediaInfoService := services.ConstructMediaInfoService(mediainfo.NewMediaInfo())

	//Setup Course Services
	courseService := services.ConstructCourseService(&dbRepository, &storageService, &mediaInfoService)

	//Setup Course Devlivery/Http Controller
	controllers.SetupCourseHandler(engine, courseService)

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
