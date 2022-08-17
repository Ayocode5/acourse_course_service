package controllers

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/http/requests"
	"acourse-course-service/pkg/http/response"
	"acourse-course-service/pkg/models"
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"strconv"
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

	page, ok := c.GetQuery("page")
	if page == "" || !ok {
		page = "1"
	}

	q_page, err2 := strconv.ParseInt(page, 10, 64)
	if err2 != nil {
		return
	}

	pagination := models.Pagination{
		Page:    q_page,
		PerPage: 25,
	}

	courses, err := hanlder.CourseService.Fetch(hanlder.Context, excludedField, pagination)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response.HttpPaginationResponse{
		HttpResponse: response.HttpResponse{
			Data:       courses,
			StatusCode: 200,
		},
		PerPage: pagination.PerPage,
		Page:    pagination.Page,
	})
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

func (hanlder CourseHanlder) DeleteCourse(c *gin.Context) {

}

func (hanlder CourseHanlder) DeleteMaterial(c *gin.Context) {

}
