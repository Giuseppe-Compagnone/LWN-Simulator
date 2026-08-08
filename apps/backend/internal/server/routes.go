package server

import (
	"github.com/gin-gonic/gin"

	"lwn-simulator-backend/internal/handlers"
	"lwn-simulator-backend/internal/services"
)

type Services struct {
	Device *services.DeviceService
}

func registerRoutes(r *gin.Engine, port string, services Services) {

	api := r.Group("/api")

	api.GET("/info", handlers.Info)

	api.GET("/status", handlers.Status(port))

	device := api.Group("/device")

	deviceHandler := handlers.NewDeviceHandler(services.Device)

	device.POST("/create-device", deviceHandler.CreateDevice)
}
