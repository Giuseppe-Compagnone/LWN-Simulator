package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func bindJSONAndValidate[T any](
	c *gin.Context,
	validator *validator.Validate,
	req *T,
) bool {

	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return false
	}

	if err := validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return false
	}

	return true
}

func bindUriAndValidate[T any](
	c *gin.Context,
	validator *validator.Validate,
	req *T,
) bool {

	if err := c.ShouldBindUri(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return false
	}

	if err := validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return false
	}

	return true
}
