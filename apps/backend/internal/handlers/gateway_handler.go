package handlers

import (
	"lwn-simulator-backend/internal/services"
	"net/http"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type GatewayHandler struct {
	service   *services.GatewayService
	validator *validator.Validate
}

func NewGatewayHandler(
	service *services.GatewayService,
	validator *validator.Validate,
) *GatewayHandler {
	return &GatewayHandler{
		service:   service,
		validator: validator,
	}
}

func (h *GatewayHandler) CreateGateway(c *gin.Context) {
	var req contracts.CreateGatewayRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	gateway, err := h.service.CreateGateway(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, contracts.CreateGatewayResponse{
		Gateway: gateway,
	})
}
