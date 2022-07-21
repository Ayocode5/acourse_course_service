package main

import (
	"acourse-course-service/pkg/database"
	"acourse-course-service/pkg/http/controllers"
	repositories "acourse-course-service/pkg/repositories/database"
	"acourse-course-service/pkg/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
)

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

	//Setup Course DBRepository
	courseRepository := repositories.ConstructDBRepository(mongodb.GetConnection(), mongodb.GetCollection())
	//Setup Course Services
	courseService := services.ConstructCourseService(courseRepository)
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
