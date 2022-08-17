package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
)

type Authorization struct {
	Permission string
	UserID     string
	Role       string
}

func AuthorizeRequestMiddleware(c *gin.Context) {

	userPermission := c.Request.Header.Get("X-User-Permission")
	userRole := c.Request.Header.Get("X-User-Role")
	userId := c.Request.Header.Get("X-User-Id")

	if userId != "" && userRole != "" && userPermission != "" {

		log.Println("Request Authorized")

		c.Set("authorization", &Authorization{
			Permission: userPermission,
			UserID:     userId,
			Role:       userRole,
		})
		c.Next()

	} else {
		log.Println("Request Unauthorized")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Request Unathorized"})
		c.Abort()
	}

}

func CanCreateCourseMiddleware(c *gin.Context) {

	permission, _ := c.Get("authorization")
	p := permission.(*Authorization).Permission

	if !strings.Contains(p, "c") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Can't Create, Unathorized"})
		c.Abort()
	}
	c.Next()
}

func CanUpdateCourseMiddleware(c *gin.Context) {

	permission, _ := c.Get("authorization")
	p := permission.(*Authorization).Permission

	if !strings.Contains(p, "u") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Can't Update, Unathorized"})
		c.Abort()
	}
	c.Next()
}

func CanDeleteCourseMiddleware(c *gin.Context) {

	permission, _ := c.Get("authorization")
	p := permission.(*Authorization).Permission

	if !strings.Contains(p, "d") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Can't Delete, Unathorized"})
		c.Abort()
	}
	c.Next()
}
