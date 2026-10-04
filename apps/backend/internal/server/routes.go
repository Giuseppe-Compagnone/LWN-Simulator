package server

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"lwn-simulator-backend/internal/handlers"
	"lwn-simulator-backend/internal/services"
)

type Services struct {
	Device     *services.DeviceService
	Gateway    *services.GatewayService
	Simulation *services.SimulationService
}

func registerRoutes(r *gin.Engine, port string, services Services) {

	validator := validator.New()

	api := r.Group("/api")

	appInfo := api.Group("/app-info")

	appInfoHandler := handlers.NewAppInfoHandler()

	appInfo.GET("/info", appInfoHandler.Info)
	appInfo.GET("/status", appInfoHandler.Status(port))

	device := api.Group("/device")

	deviceHandler := handlers.NewDeviceHandler(services.Device, validator)

	device.POST("/create-device", deviceHandler.CreateDevice)
	device.GET("/get-device/:id", deviceHandler.GetDevice)
	device.GET("/get-devices", deviceHandler.GetDevices)
	device.PUT("/update-device/:id", deviceHandler.UpdateDevice)
	device.DELETE("/delete-device/:id", deviceHandler.DeleteDevice)

	gateway := api.Group("/gateway")

	gatewayHandler := handlers.NewGatewayHandler(services.Gateway, validator)

	gateway.POST("/create-gateway", gatewayHandler.CreateGateway)
	gateway.GET("/get-gateway/:id", gatewayHandler.GetGateway)
	gateway.GET("/get-gateways", gatewayHandler.GetGateways)
	gateway.PUT("/update-gateway/:id", gatewayHandler.UpdateGateway)
	gateway.DELETE("/delete-gateway/:id", gatewayHandler.DeleteGateway)

	simulation := api.Group("/simulation")
	simulationHandler := handlers.NewSimulationHandler(services.Simulation, validator)
	simulation.POST("/start", simulationHandler.Start)
	simulation.POST("/pause", simulationHandler.Pause)
	simulation.POST("/resume", simulationHandler.Resume)
	simulation.POST("/stop", simulationHandler.Stop)
	simulation.GET("/snapshot", simulationHandler.Snapshot)
	simulation.GET("/events", simulationHandler.Events)
	simulation.GET("/metrics", simulationHandler.Metrics)
	simulation.POST("/uplinks", simulationHandler.QueueUplink)
	simulation.POST("/downlinks", simulationHandler.QueueDownlink)
	simulation.POST("/mac-commands", simulationHandler.QueueMACCommand)
	simulation.GET("/ws", simulationHandler.WebSocket)
}
