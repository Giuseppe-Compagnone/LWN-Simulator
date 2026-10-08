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
	realtimeHub := realtime.NewHub()
	udpOptions := gateway.UDPOptions{LocalAddress: os.Getenv("LWN_GATEWAY_UDP_LISTEN_ADDR")}
	profileRuntimeManager := services.NewProfileRuntimeManager(
		dataDir,
		profileService,
		types.Options{GatewayAdapterFactory: gateway.NewUDPFactory(udpOptions)},
		realtimeHub,
	)
	profileService.SetRealtimePublisher(realtimeHub)
	defaultRuntime, err := profileRuntimeManager.Runtime(defaultProfile.ID)
	if err != nil {
		return nil, fmt.Errorf("initialize default profile runtime: %w", err)
	}
	if err := profileRuntimeManager.RestorePersisted(context.Background()); err != nil {
		return nil, fmt.Errorf("restore persisted simulation: %w", err)
	}

	registerMiddleware(r)

	registerRoutes(r, port, Services{
		Device:     defaultRuntime.Device,
		Gateway:    defaultRuntime.Gateway,
		Profile:    profileService,
		ProfileID:  defaultProfile.ID,
		Simulation: defaultRuntime.Simulation,
		Profiles:   profileRuntimeManager,
		Realtime:   handlers.NewRealtimeHandler(realtimeHub),
	})

	if err := registerFrontend(r); err != nil {
		return nil, err
	}

	return r, nil
}
