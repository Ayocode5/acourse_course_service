package controllers

import (
	"acourse-course-service/pkg/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

func ListCourse(c *gin.Context) {
	c.JSON(http.StatusOK, models.AllCourse())
}

func FindCourse(c *gin.Context) {

	course, err := models.FindCourse(c.Param("id"))
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

func CreateCourse() {

}
