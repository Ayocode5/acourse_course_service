package main

import (
	"acourse-course-service/pkg/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	engine := gin.Default()
	Routes.RegisterRoutes(engine)

	if port := os.Getenv("APP_PORT"); port == "" {
		engine.Run(":8080")
	} else {
		engine.Run(":" + port)
	}

}
