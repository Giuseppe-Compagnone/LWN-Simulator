package handlers

import (
	"net/http"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/gin-gonic/gin"
)

func Status(c *gin.Context) {

	res := contracts.Response{
		Time: time.Now().Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, res)
}
