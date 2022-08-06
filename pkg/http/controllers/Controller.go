package controllers

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupCourseHandler(router *gin.Engine, courseService contracts.CourseService) {

	handler := &CourseHanlder{CourseService: courseService}

	r := router.Group("/course/")
	r.Use(middleware.AuthorizeRequestMiddleware)
	r.GET("/all", handler.ListCourse)
	r.GET("/show/:id", handler.FindCourse)
	r.POST("/create", handler.CreateCourse)

}
