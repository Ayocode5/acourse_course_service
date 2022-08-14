package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Authorization struct {
	Permission string
	UserID     string
	Role       string
}

func AuthorizeRequestMiddleware(c *gin.Context) {

	log.Println("[!] Acourse Authorize Middleware Hit [!]")

	//log.Println(c.Request.Header.Get("X-User-Permission"))

	userPermission := c.Request.Header.Get("X-User-Permission")
	userRole := c.Request.Header.Get("X-User-Role")
	userId := c.Request.Header.Get("X-User-Id")

	if userId != "" && userRole != "" && userPermission != "" {
		c.Set("authorization", &Authorization{
			Permission: userPermission,
			UserID:     userId,
			Role:       userRole,
		})
		c.Next()
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unathorized"})
		c.Abort()
	}

}
