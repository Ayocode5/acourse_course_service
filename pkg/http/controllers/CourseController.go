package controllers

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/http/requests"
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"strings"
)

type CourseHanlder struct {
	CourseService contracts.CourseService
	Context       context.Context
}

func (hanlder *CourseHanlder) FetchAll(c *gin.Context) {

	excludedField := []string{}
	if c.Query("exclude") != "" {
		excludedField = strings.Split(c.Query("exclude"), ",")
	}

	courses, err := hanlder.CourseService.Fetch(hanlder.Context, excludedField)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, courses)
}

func (handler *CourseHanlder) Find(c *gin.Context) {

	excludedField := []string{}
	if c.Query("exclude") != "" {
		excludedField = strings.Split(c.Query("exclude"), ",")
	}

	course, err := handler.CourseService.FetchById(handler.Context, c.Param("id"), excludedField)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, course)
}

func (handler *CourseHanlder) CreateCourse(c *gin.Context) {

	//Validate Request
	var createCourseRequest requests.CreateCourseRequest

	err := c.ShouldBind(&createCourseRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := handler.CourseService.Create(handler.Context, createCourseRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Course Successfuly Created", "data": res})
	return
}

func (hanlder CourseHanlder) UpdateCourse(c *gin.Context) {

	val, _ := c.Get("authorization")
	authContext := context.WithValue(hanlder.Context, "authorization", val)

	//Validate Request
	var updateCourseRequest requests.UpdateCourseRequest

	err := c.ShouldBind(&updateCourseRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := hanlder.CourseService.Update(authContext, updateCourseRequest, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(res.StatusCode, res)
	return
}
