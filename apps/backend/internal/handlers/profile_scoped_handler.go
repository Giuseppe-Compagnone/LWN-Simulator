package handlers

import (
	"net/http"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"lwn-simulator-backend/internal/services"
)

type ProfileRuntimeResolver interface {
	Runtime(string) (*services.ProfileRuntime, error)
}

type ProfileScopedHandler struct {
	resolver  ProfileRuntimeResolver
	validator *validator.Validate
}

func NewProfileScopedHandler(resolver ProfileRuntimeResolver, validator *validator.Validate) *ProfileScopedHandler {
	return &ProfileScopedHandler{resolver: resolver, validator: validator}
}

func (h *ProfileScopedHandler) runtime(c *gin.Context) (*services.ProfileRuntime, bool) {
	runtime, err := h.resolver.Runtime(c.Param("profileID"))
	if err != nil {
		writeServiceError(c, err)
		return nil, false
	}
	return runtime, true
}

func (h *ProfileScopedHandler) CreateDevice(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewDeviceHandler(runtime.Device, h.validator).CreateDevice(c)
	}
}

func (h *ProfileScopedHandler) GetDevice(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewDeviceHandler(runtime.Device, h.validator).GetDevice(c)
	}
}

func (h *ProfileScopedHandler) GetDevices(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewDeviceHandler(runtime.Device, h.validator).GetDevices(c)
	}
}

func (h *ProfileScopedHandler) UpdateDevice(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewDeviceHandler(runtime.Device, h.validator).UpdateDevice(c)
	}
}

func (h *ProfileScopedHandler) DeleteDevice(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewDeviceHandler(runtime.Device, h.validator).DeleteDevice(c)
	}
}

func (h *ProfileScopedHandler) CreateGateway(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewGatewayHandler(runtime.Gateway, h.validator).CreateGateway(c)
	}
}

func (h *ProfileScopedHandler) GetGateway(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewGatewayHandler(runtime.Gateway, h.validator).GetGateway(c)
	}
}

func (h *ProfileScopedHandler) GetGateways(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewGatewayHandler(runtime.Gateway, h.validator).GetGateways(c)
	}
}

func (h *ProfileScopedHandler) UpdateGateway(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewGatewayHandler(runtime.Gateway, h.validator).UpdateGateway(c)
	}
}

func (h *ProfileScopedHandler) DeleteGateway(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewGatewayHandler(runtime.Gateway, h.validator).DeleteGateway(c)
	}
}

func (h *ProfileScopedHandler) GetGatewayBridge(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		config, err := runtime.Configuration.GetGatewayBridge()
		if err != nil {
			writeServiceError(c, err)
			return
		}
		c.JSON(http.StatusOK, config)
	}
}

func (h *ProfileScopedHandler) UpdateGatewayBridge(c *gin.Context) {
	var config contracts.GatewayBridgeConfig
	if !bindJSONAndValidate(c, h.validator, &config) {
		return
	}
	runtime, ok := h.runtime(c)
	if !ok {
		return
	}
	updated, err := runtime.Configuration.UpdateGatewayBridge(config)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *ProfileScopedHandler) StartSimulation(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).Start(c)
	}
}

func (h *ProfileScopedHandler) SetSimulationSpeed(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).SetSpeed(c)
	}
}

func (h *ProfileScopedHandler) PauseSimulation(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).Pause(c)
	}
}

func (h *ProfileScopedHandler) ResumeSimulation(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).Resume(c)
	}
}

func (h *ProfileScopedHandler) StopSimulation(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).Stop(c)
	}
}

func (h *ProfileScopedHandler) SimulationSnapshot(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).Snapshot(c)
	}
}

func (h *ProfileScopedHandler) SimulationEvents(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).Events(c)
	}
}

func (h *ProfileScopedHandler) SimulationLogs(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).Logs(c)
	}
}

func (h *ProfileScopedHandler) SimulationLog(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).Log(c)
	}
}

func (h *ProfileScopedHandler) SimulationLogEvents(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).LogEvents(c)
	}
}

func (h *ProfileScopedHandler) SimulationMetrics(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).Metrics(c)
	}
}

func (h *ProfileScopedHandler) QueueUplink(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).QueueUplink(c)
	}
}

func (h *ProfileScopedHandler) QueueDownlink(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).QueueDownlink(c)
	}
}

func (h *ProfileScopedHandler) QueueMACCommand(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).QueueMACCommand(c)
	}
}

func (h *ProfileScopedHandler) SimulationWebSocket(c *gin.Context) {
	if runtime, ok := h.runtime(c); ok {
		NewSimulationHandlerForProfile(runtime.Simulation, h.validator, runtime.Profile.ID).WebSocket(c)
	}
}
