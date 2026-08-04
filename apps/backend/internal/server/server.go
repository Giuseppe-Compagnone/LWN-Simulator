package server

import (
	"lwn-simulator-backend/internal/frontend"
	"net/http"

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

func New() *gin.Engine {

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	registerMiddleware(r)

	registerRoutes(r)

	registerFrontend(r)

	return r
}
