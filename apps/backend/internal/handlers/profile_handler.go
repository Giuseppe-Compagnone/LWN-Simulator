package handlers

import (
	"net/http"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"lwn-simulator-backend/internal/services"
)

type ProfileService interface {
	List() ([]contracts.Profile, error)
	Get(string) (contracts.Profile, error)
	Create(string) (contracts.Profile, error)
	Rename(string, string) (contracts.Profile, error)
	Delete(string) error
}

type ProfileHandler struct {
	service   ProfileService
	validator *validator.Validate
}

func NewProfileHandler(service ProfileService, validator *validator.Validate) *ProfileHandler {
	return &ProfileHandler{service: service, validator: validator}
}

func (h *ProfileHandler) GetProfiles(c *gin.Context) {
	profiles, err := h.service.List()
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, contracts.GetProfilesResponse{Profiles: profiles, Total: int32(len(profiles))})
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	profile, err := h.service.Get(c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, contracts.GetProfileResponse{Profile: profile})
}

func (h *ProfileHandler) CreateProfile(c *gin.Context) {
	var req contracts.CreateProfileRequest
	if !bindJSONAndValidate(c, h.validator, &req) {
		return
	}
	profile, err := h.service.Create(req.Name)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, contracts.CreateProfileResponse{Profile: profile})
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	var req contracts.UpdateProfileRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request path"})
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	profile, err := h.service.Rename(req.ID, req.Name)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, contracts.UpdateProfileResponse{Profile: profile})
}

func (h *ProfileHandler) DeleteProfile(c *gin.Context) {
	if err := h.service.Delete(c.Param("id")); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

var _ ProfileService = (*services.ProfileService)(nil)
