package server

import (
	"lwn-simulator-backend/internal/frontend"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func registerFrontend(r *gin.Engine) {

	webFS, err := frontend.Files()
	if err != nil {
		panic(err)
	}

	r.NoRoute(gin.WrapH(
		http.FileServer(http.FS(webFS)),
	))
}

func New(port string) *gin.Engine {

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
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

	registerMiddleware(r)

	registerRoutes(r, port)

	registerFrontend(r)

	return r
}
