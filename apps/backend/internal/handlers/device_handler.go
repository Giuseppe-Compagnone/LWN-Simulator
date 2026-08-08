package handlers

import (
	"net/http"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"lwn-simulator-backend/internal/services"
)

type DeviceHandler struct {
	service   *services.DeviceService
	validator *validator.Validate
}

func NewDeviceHandler(
	service *services.DeviceService,
	validator *validator.Validate,
) *DeviceHandler {
	return &DeviceHandler{
		service:   service,
		validator: validator,
	}
}

func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	var req contracts.CreateDeviceRequest

	if !bindAndValidate(c, h.validator, &req) {
		return
	}

	res, err := h.service.CreateDevice(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *DeviceHandler) GetDevice(c *gin.Context) {
	var req contracts.GetDeviceRequest

	if !bindAndValidate(c, h.validator, &req) {
		return
	}

	res, err := h.service.GetDevice(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, res)
}
