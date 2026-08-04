package server

import (
	"github.com/gin-gonic/gin"

	"lwn-simulator-backend/internal/handlers"
)

func registerRoutes(r *gin.Engine) {

	api := r.Group("/api")

	api.GET("/info", handlers.Info)

	api.GET("/status", handlers.Status)

}
