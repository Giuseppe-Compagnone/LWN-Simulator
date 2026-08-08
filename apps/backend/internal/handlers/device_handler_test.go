package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type mockDeviceService struct {
	createDeviceResponse contracts.CreateDeviceResponse
	getDeviceResponse    contracts.GetDeviceResponse
	getDevicesResponse   contracts.GetDevicesResponse
	updateDeviceResponse contracts.UpdateDeviceResponse
	deleteDeviceResponse contracts.DeleteDeviceResponse

	createDeviceErr error
	getDeviceErr    error
	getDevicesErr   error
	updateDeviceErr error
	deleteDeviceErr error

	createDeviceCalls int
	getDeviceCalls    int
	getDevicesCalls   int
	updateDeviceCalls int
	deleteDeviceCalls int

	lastCreateRequest contracts.CreateDeviceRequest
	lastGetRequest    contracts.GetDeviceRequest
	lastGetDevicesReq contracts.GetDevicesRequest
	lastUpdateRequest contracts.UpdateDeviceRequest
	lastDeleteRequest contracts.DeleteDeviceRequest
}

const (
	testDeviceID1 = "550e8400-e29b-41d4-a716-446655440000"
	testDeviceID2 = "6ba7b810-9dad-41d1-80b4-00c04fd430c8"
)

func (m *mockDeviceService) CreateDevice(
	req contracts.CreateDeviceRequest,
) (contracts.CreateDeviceResponse, error) {
	m.createDeviceCalls++
	m.lastCreateRequest = req

	return m.createDeviceResponse, m.createDeviceErr
}

func (m *mockDeviceService) GetDevice(
	req contracts.GetDeviceRequest,
) (contracts.GetDeviceResponse, error) {
	m.getDeviceCalls++
	m.lastGetRequest = req

	return m.getDeviceResponse, m.getDeviceErr
}

func (m *mockDeviceService) GetDevices(
	req contracts.GetDevicesRequest,
) (contracts.GetDevicesResponse, error) {
	m.getDevicesCalls++
	m.lastGetDevicesReq = req

	return m.getDevicesResponse, m.getDevicesErr
}

func (m *mockDeviceService) UpdateDevice(
	req contracts.UpdateDeviceRequest,
) (contracts.UpdateDeviceResponse, error) {
	m.updateDeviceCalls++
	m.lastUpdateRequest = req

	return m.updateDeviceResponse, m.updateDeviceErr
}

func (m *mockDeviceService) DeleteDevice(
	req contracts.DeleteDeviceRequest,
) (contracts.DeleteDeviceResponse, error) {
	m.deleteDeviceCalls++
	m.lastDeleteRequest = req

	return m.deleteDeviceResponse, m.deleteDeviceErr
}

func newTestDeviceHandler(
	service *mockDeviceService,
) *DeviceHandler {
	return NewDeviceHandler(
		service,
		validator.New(),
	)
}

func setupRouter(handler *DeviceHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.POST("/devices", handler.CreateDevice)
	router.GET("/devices/:id", handler.GetDevice)
	router.GET("/devices", handler.GetDevices)
	router.PUT("/devices/:id", handler.UpdateDevice)
	router.DELETE("/devices/:id", handler.DeleteDevice)

	return router
}

func TestDeviceHandler_CreateDevice(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		service        *mockDeviceService
		wantStatusCode int
		wantCalls      int
	}{
		{
			name: "creates device",
			body: `{
				"devEUI": "0102030405060708",
				"latitude": 37.5,
				"longitude": 15.0
			}`,
			service: &mockDeviceService{
				createDeviceResponse: contracts.CreateDeviceResponse{
					Device: contracts.Device{
						ID:        testDeviceID1,
						DevEUI:    "0102030405060708",
						Latitude:  37.5,
						Longitude: 15.0,
					},
				},
			},
			wantStatusCode: http.StatusCreated,
			wantCalls:      1,
		},
		{
			name: "returns bad request for invalid JSON",
			body: `{
				"devEUI": "invalid"
			`,
			service:        &mockDeviceService{},
			wantStatusCode: http.StatusBadRequest,
			wantCalls:      0,
		},
		{
			name: "returns bad request for validation error",
			body: `{
				"devEUI": "0102030405060708",
				"latitude": 100,
				"longitude": 15
			}`,
			service:        &mockDeviceService{},
			wantStatusCode: http.StatusBadRequest,
			wantCalls:      0,
		},
		{
			name: "returns internal server error",
			body: `{
				"devEUI": "0102030405060708",
				"latitude": 37.5,
				"longitude": 15
			}`,
			service: &mockDeviceService{
				createDeviceErr: errors.New("database error"),
			},
			wantStatusCode: http.StatusInternalServerError,
			wantCalls:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestDeviceHandler(tt.service)
			router := setupRouter(handler)

			req := httptest.NewRequest(
				http.MethodPost,
				"/devices",
				strings.NewReader(tt.body),
			)

			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatusCode {
				t.Fatalf(
					"status code = %d, want %d",
					recorder.Code,
					tt.wantStatusCode,
				)
			}

			if tt.service.createDeviceCalls != tt.wantCalls {
				t.Errorf(
					"CreateDevice() calls = %d, want %d",
					tt.service.createDeviceCalls,
					tt.wantCalls,
				)
			}
		})
	}
}

func TestDeviceHandler_GetDevice(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		service        *mockDeviceService
		wantStatusCode int
		wantCalls      int
	}{
		{
			name: "returns device",
			id:   testDeviceID1,
			service: &mockDeviceService{
				getDeviceResponse: contracts.GetDeviceResponse{
					Device: contracts.Device{
						ID:        testDeviceID1,
						DevEUI:    "0102030405060708",
						Latitude:  37.5,
						Longitude: 15.0,
					},
				},
			},
			wantStatusCode: http.StatusOK,
			wantCalls:      1,
		},
		{
			name: "returns error when service fails",
			id:   testDeviceID1,
			service: &mockDeviceService{
				getDeviceErr: errors.New("device not found"),
			},
			wantStatusCode: http.StatusInternalServerError,
			wantCalls:      1,
		},
		{
			name:           "returns bad request for invalid id",
			id:             "invalid",
			service:        &mockDeviceService{},
			wantStatusCode: http.StatusBadRequest,
			wantCalls:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestDeviceHandler(tt.service)
			router := setupRouter(handler)

			req := httptest.NewRequest(
				http.MethodGet,
				"/devices/"+tt.id,
				nil,
			)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatusCode {
				t.Fatalf(
					"status code = %d, want %d",
					recorder.Code,
					tt.wantStatusCode,
				)
			}

			if tt.service.getDeviceCalls != tt.wantCalls {
				t.Errorf(
					"GetDevice() calls = %d, want %d",
					tt.service.getDeviceCalls,
					tt.wantCalls,
				)
			}

			if tt.wantCalls > 0 &&
				tt.service.lastGetRequest.ID != tt.id {
				t.Errorf(
					"GetDevice() ID = %q, want %q",
					tt.service.lastGetRequest.ID,
					tt.id,
				)
			}
		})
	}
}

func TestDeviceHandler_GetDevices(t *testing.T) {
	tests := []struct {
		name           string
		service        *mockDeviceService
		wantStatusCode int
		wantCalls      int
	}{
		{
			name: "returns devices",
			service: &mockDeviceService{
				getDevicesResponse: contracts.GetDevicesResponse{
					Devices: &[]contracts.Device{
						{
							ID:     testDeviceID1,
							DevEUI: "0102030405060708",
						},
						{
							ID:     testDeviceID2,
							DevEUI: "1122334455667788",
						},
					},
				},
			},
			wantStatusCode: http.StatusOK,
			wantCalls:      1,
		},
		{
			name: "returns internal server error",
			service: &mockDeviceService{
				getDevicesErr: errors.New("database error"),
			},
			wantStatusCode: http.StatusInternalServerError,
			wantCalls:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestDeviceHandler(tt.service)
			router := setupRouter(handler)

			req := httptest.NewRequest(
				http.MethodGet,
				"/devices",
				nil,
			)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatusCode {
				t.Fatalf(
					"status code = %d, want %d",
					recorder.Code,
					tt.wantStatusCode,
				)
			}

			if tt.service.getDevicesCalls != tt.wantCalls {
				t.Errorf(
					"GetDevices() calls = %d, want %d",
					tt.service.getDevicesCalls,
					tt.wantCalls,
				)
			}
		})
	}
}

func TestDeviceHandler_UpdateDevice(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		body           string
		service        *mockDeviceService
		wantStatusCode int
		wantCalls      int
	}{
		{
			name: "updates device",
			id:   testDeviceID1,
			body: `{
				"devEUI": "1122334455667788",
				"latitude": 40,
				"longitude": 20
			}`,
			service: &mockDeviceService{
				updateDeviceResponse: contracts.UpdateDeviceResponse{
					Device: contracts.Device{
						ID:        testDeviceID1,
						DevEUI:    "1122334455667788",
						Latitude:  40,
						Longitude: 20,
					},
				},
			},
			wantStatusCode: http.StatusOK,
			wantCalls:      1,
		},
		{
			name: "updates partial device",
			id:   testDeviceID1,
			body: `{
				"latitude": 40
			}`,
			service: &mockDeviceService{
				updateDeviceResponse: contracts.UpdateDeviceResponse{
					Device: contracts.Device{
						ID:       testDeviceID1,
						Latitude: 40,
					},
				},
			},
			wantStatusCode: http.StatusOK,
			wantCalls:      1,
		},
		{
			name: "returns bad request for invalid body",
			id:   testDeviceID1,
			body: `{
				"latitude": 100
			}`,
			service:        &mockDeviceService{},
			wantStatusCode: http.StatusBadRequest,
			wantCalls:      0,
		},
		{
			name: "returns internal server error",
			id:   testDeviceID1,
			body: `{
				"latitude": 40
			}`,
			service: &mockDeviceService{
				updateDeviceErr: errors.New("database error"),
			},
			wantStatusCode: http.StatusInternalServerError,
			wantCalls:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestDeviceHandler(tt.service)
			router := setupRouter(handler)

			req := httptest.NewRequest(
				http.MethodPut,
				"/devices/"+tt.id,
				strings.NewReader(tt.body),
			)

			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatusCode {
				t.Fatalf(
					"status code = %d, want %d",
					recorder.Code,
					tt.wantStatusCode,
				)
			}

			if tt.service.updateDeviceCalls != tt.wantCalls {
				t.Errorf(
					"UpdateDevice() calls = %d, want %d",
					tt.service.updateDeviceCalls,
					tt.wantCalls,
				)
			}

			if tt.wantCalls > 0 {
				req := tt.service.lastUpdateRequest

				if req.ID != tt.id {
					t.Errorf(
						"UpdateDevice() ID = %q, want %q",
						req.ID,
						tt.id,
					)
				}
			}
		})
	}
}

func TestDeviceHandler_DeleteDevice(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		service        *mockDeviceService
		wantStatusCode int
		wantCalls      int
	}{
		{
			name: "deletes device",
			id:   testDeviceID1,
			service: &mockDeviceService{
				deleteDeviceResponse: contracts.DeleteDeviceResponse{},
			},
			wantStatusCode: http.StatusNoContent,
			wantCalls:      1,
		},
		{
			name: "returns internal server error",
			id:   testDeviceID1,
			service: &mockDeviceService{
				deleteDeviceErr: errors.New("database error"),
			},
			wantStatusCode: http.StatusInternalServerError,
			wantCalls:      1,
		},
		{
			name:           "returns bad request for invalid id",
			id:             "invalid",
			service:        &mockDeviceService{},
			wantStatusCode: http.StatusBadRequest,
			wantCalls:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestDeviceHandler(tt.service)
			router := setupRouter(handler)

			req := httptest.NewRequest(
				http.MethodDelete,
				"/devices/"+tt.id,
				nil,
			)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatusCode {
				t.Fatalf(
					"status code = %d, want %d",
					recorder.Code,
					tt.wantStatusCode,
				)
			}

			if tt.service.deleteDeviceCalls != tt.wantCalls {
				t.Errorf(
					"DeleteDevice() calls = %d, want %d",
					tt.service.deleteDeviceCalls,
					tt.wantCalls,
				)
			}

			if tt.wantCalls > 0 &&
				tt.service.lastDeleteRequest.ID != tt.id {
				t.Errorf(
					"DeleteDevice() ID = %q, want %q",
					tt.service.lastDeleteRequest.ID,
					tt.id,
				)
			}
		})
	}
}
