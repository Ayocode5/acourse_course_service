package Routes

import (
	"acourse-course-service/pkg/http/controllers"
	"acourse-course-service/pkg/http/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {

	r := router.Group("/course/")
	r.Use(middleware.AuthorizeRequestMiddleware)
	r.GET("/all", controllers.ListCourse)
	r.GET("/show/:id", controllers.FindCourse)

}
