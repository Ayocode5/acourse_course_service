package controllers

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/http/middleware"
	"context"
	"github.com/gin-gonic/gin"
)

func SetupCourseHandler(ctx context.Context, router *gin.Engine, courseService contracts.CourseService) {

	handler := &CourseHanlder{CourseService: courseService, Context: ctx}

	r := router.Group("/course/")
	r.Use(middleware.AuthorizeRequestMiddleware)
	r.GET("/list", handler.FetchAll)
	r.GET("/show/:id", handler.Find)
	r.POST("/create", handler.CreateCourse)
	r.PUT("/update/:id", handler.UpdateCourse)

}
