package server

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesIncludesSimulationRuntimeAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerRoutes(router, "8080", Services{})
	routes := make(map[string]struct{})
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}
	for _, expected := range []string{
		"POST /api/simulation/start",
		"POST /api/simulation/pause",
		"POST /api/simulation/resume",
		"POST /api/simulation/stop",
		"GET /api/simulation/snapshot",
		"GET /api/simulation/events",
		"GET /api/simulation/logs",
		"GET /api/simulation/logs/:id",
		"GET /api/simulation/logs/:id/events",
		"GET /api/simulation/metrics",
		"POST /api/simulation/uplinks",
		"POST /api/simulation/downlinks",
		"POST /api/simulation/mac-commands",
		"GET /api/simulation/ws",
	} {
		if _, exists := routes[expected]; !exists {
			t.Errorf("route %q is not registered", expected)
		}
	}
}
