package handlers

import (
	"net/http"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/gin-gonic/gin"

	"lwn-simulator-backend/internal/services"
)

type DeviceHandler struct {
	service *services.DeviceService
}

func NewDeviceHandler(service *services.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		service: service,
	}
}

func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	var req contracts.CreateDeviceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	device, err := h.service.CreateDevice(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, contracts.CreateDeviceResponse{
		Device: device,
	})
}
