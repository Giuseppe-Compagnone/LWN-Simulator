package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type profileHandlerProfileServiceStub struct{}

func (profileHandlerProfileServiceStub) List() ([]contracts.Profile, error) {
	return nil, nil
}

func (profileHandlerProfileServiceStub) Get(string) (contracts.Profile, error) {
	return contracts.Profile{}, nil
}

func (profileHandlerProfileServiceStub) Create(string) (contracts.Profile, error) {
	return contracts.Profile{}, nil
}

func (profileHandlerProfileServiceStub) Rename(string, string) (contracts.Profile, error) {
	return contracts.Profile{}, nil
}

func (profileHandlerProfileServiceStub) Delete(string) error {
	return nil
}

type profileHandlerArchiveServiceStub struct{}

func (profileHandlerArchiveServiceStub) ExportArchive(string) (contracts.ProfileArchive, error) {
	return contracts.ProfileArchive{}, nil
}

func (profileHandlerArchiveServiceStub) ImportArchive(contracts.ProfileArchive) (contracts.Profile, error) {
	return contracts.Profile{}, nil
}

func TestExportProfileMarksDownloadResponsesAsAttachments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewProfileHandler(
		profileHandlerProfileServiceStub{},
		validator.New(),
		profileHandlerArchiveServiceStub{},
	)
	router.GET("/profiles/:id/export", handler.ExportProfile)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/profiles/profile-id/export?download=1",
		nil,
	)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Disposition"); got != `attachment; filename="lwn-profile.json"` {
		t.Fatalf("expected attachment header, got %q", got)
	}
}
