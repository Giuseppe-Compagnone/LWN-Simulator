package handlers

import (
	"lwn-simulator-backend/version"
	"net/http"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/gin-gonic/gin"
)

func Info(c *gin.Context) {

	res := contracts.AppInfoResponse{
		App:     "lwn-simulator",
		Version: version.AppVersion,
	}

	c.JSON(http.StatusOK, res)
}
