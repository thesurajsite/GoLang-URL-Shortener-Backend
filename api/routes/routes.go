package routes

import (
	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {
	r.GET("/:url", ResolveURL)
	r.POST("/api/v1/", ShortenURL)
}
