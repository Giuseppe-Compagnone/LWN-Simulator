package handlers

import (
	"net/http"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/gin-gonic/gin"
)

func Status(port string) gin.HandlerFunc {
	return func(c *gin.Context) {

		res := contracts.StatusResponse{
			Port: port,
		}

		c.JSON(http.StatusOK, res)
	}
}
