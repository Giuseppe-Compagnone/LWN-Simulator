package handlers

import (
	"context"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type SimulationLogExportService interface {
	ExportLogs(context.Context, string, io.Writer) error
}

func (h *SimulationHandler) ExportLogs(c *gin.Context) {
	service, ok := h.service.(SimulationLogExportService)
	if !ok {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "simulation log export is not configured"})
		return
	}
	file, err := os.CreateTemp("", "lwn-logs-*")
	if err != nil {
		writeServiceError(c, err)
		return
	}
	defer os.Remove(file.Name())
	defer file.Close()
	id := c.Param("id")
	if err := service.ExportLogs(c.Request.Context(), id, file); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "simulation log not found"})
			return
		}
		writeServiceError(c, err)
		return
	}
	filename, contentType := "simulation-logs.zip", "application/zip"
	if id != "" {
		filename, contentType = "simulation-"+url.PathEscape(id)+".json", "application/json"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	c.Header("Cache-Control", "no-store")
	http.ServeContent(c.Writer, c.Request, filename, time.Time{}, file)
}
