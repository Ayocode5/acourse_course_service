package controllers

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/http/requests"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

type CourseHanlder struct {
	CourseService contracts.CourseService
}

func (hanlder *CourseHanlder) ListCourse(c *gin.Context) {
	courses, err := hanlder.CourseService.Fetch()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, courses)
}

func (handler *CourseHanlder) FindCourse(c *gin.Context) {

	course, err := handler.CourseService.FetchById(c.Param("id"))

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

	res, err := handler.CourseService.Create(createCourseRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success", "data": res})
	return
}
