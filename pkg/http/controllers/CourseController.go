package controllers

import (
	"acourse-course-service/pkg/contracts"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

type CourseHanlder struct {
	Service contracts.CourseService
}

func (hanlder *CourseHanlder) ListCourse(c *gin.Context) {
	courses, err := hanlder.Service.Fetch()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, courses)
}

func (handler *CourseHanlder) FindCourse(c *gin.Context) {

	course, err := handler.Service.FetchById(c.Param("id"))
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
