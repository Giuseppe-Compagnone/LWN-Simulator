package server

import (
	"context"
	"fmt"
	"lwn-simulator-backend/internal/database"
	"lwn-simulator-backend/internal/frontend"
	"lwn-simulator-backend/internal/handlers"
	"lwn-simulator-backend/internal/realtime"
	"lwn-simulator-backend/internal/repositories"
	"lwn-simulator-backend/internal/services"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Giuseppe-Compagnone/lwn-engine/gateway"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func registerFrontend(r *gin.Engine) error {

	webFS, err := frontend.Files()
	if err != nil {
		return fmt.Errorf("load embedded frontend: %w", err)
	}

	fileServer := http.FileServer(http.FS(webFS))
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	return nil
}

func New(port string) (*gin.Engine, error) {

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			if origin == "" {
				return true
			}

			parsed, err := url.Parse(origin)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
				return false
			}

			switch parsed.Hostname() {
			case "localhost", "127.0.0.1", "::1":
				return true
			default:
				return false
			}
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
	}))

	dataDir, err := database.Initialize()
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	profileRepository := repositories.NewProfileRepository(dataDir)
	profileService := services.NewProfileService(profileRepository, dataDir)
	defaultProfile, err := profileService.EnsureDefaultProfile()
	if err != nil {
		return nil, fmt.Errorf("initialize default profile: %w", err)
	}
	profileStorage, err := database.NewProfileStorage(dataDir, defaultProfile.ID)
	if err != nil {
		return nil, fmt.Errorf("resolve default profile storage: %w", err)
	}

	deviceRepository := repositories.NewDeviceRepository(profileStorage.Directory())
	deviceService := services.NewDeviceService(deviceRepository)

	gatewayRepository := repositories.NewGatewayRepository(profileStorage.Directory())
	gatewayService := services.NewGatewayService(gatewayRepository)

	udpOptions := gateway.UDPOptions{LocalAddress: os.Getenv("LWN_GATEWAY_UDP_LISTEN_ADDR")}
	simulationService := services.NewSimulationService(
		deviceService,
		gatewayService,
		types.Options{GatewayAdapterFactory: gateway.NewUDPFactory(udpOptions)},
		services.NewFileSimulationCheckpointStore(filepath.Join(profileStorage.Directory(), "simulation-checkpoint.json")),
		services.NewFileSimulationLogStore(filepath.Join(profileStorage.Directory(), "simulation-logs.json")),
	)
	realtimeHub := realtime.NewHub()
	deviceService.SetRealtimePublisher(realtimeHub)
	gatewayService.SetRealtimePublisher(realtimeHub)
	simulationService.SetRealtimePublisher(realtimeHub)
	deviceService.SetRuntimeSynchronizer(simulationService)
	gatewayService.SetRuntimeSynchronizer(simulationService)
	if err := simulationService.RestorePersisted(context.Background()); err != nil {
		return nil, fmt.Errorf("restore persisted simulation: %w", err)
	}

	registerMiddleware(r)

	registerRoutes(r, port, Services{
		Device:     deviceService,
		Gateway:    gatewayService,
		Profile:    profileService,
		Simulation: simulationService,
		Realtime:   handlers.NewRealtimeHandler(realtimeHub),
	})

	if err := registerFrontend(r); err != nil {
		return nil, err
	}

	return r, nil
}
