package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
)

func AuthorizeRequestMiddleware(c *gin.Context) {

	log.Println("[!] Acourse Authorize Middleware Hit [!]")
	log.Println(c.Request.Header)

}
