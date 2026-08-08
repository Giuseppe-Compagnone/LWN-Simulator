package server

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"lwn-simulator-backend/internal/handlers"
	"lwn-simulator-backend/internal/services"
)

type Services struct {
	Device  *services.DeviceService
	Gateway *services.GatewayService
}

func registerRoutes(r *gin.Engine, port string, services Services) {

	validator := validator.New()

	api := r.Group("/api")

	api.GET("/info", handlers.Info)
	api.GET("/status", handlers.Status(port))

	device := api.Group("/device")

	deviceHandler := handlers.NewDeviceHandler(services.Device, validator)

	device.POST("/create-device", deviceHandler.CreateDevice)
	device.GET("/get-device/:id", deviceHandler.GetDevice)
	device.GET("/get-devices", deviceHandler.GetDevices)

	gateway := api.Group("/gateway")

	gatewayHandler := handlers.NewGatewayHandler(services.Gateway, validator)

	gateway.POST("/create-gateway", gatewayHandler.CreateGateway)
}
