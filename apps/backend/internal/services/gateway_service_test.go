package services

import (
	"errors"
	"strings"
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

type mockGatewayRepository struct {
	gateways  []contracts.Gateway
	getAllErr error
	saveErr   error
}

func (m *mockGatewayRepository) GetAll() ([]contracts.Gateway, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}

	return m.gateways, nil
}

func (m *mockGatewayRepository) Save(gateways []contracts.Gateway) error {
	if m.saveErr != nil {
		return m.saveErr
	}

	m.gateways = gateways
	return nil
}

func TestGatewayService_CreateGateway(t *testing.T) {
	tests := []struct {
		name          string
		repository    *mockGatewayRepository
		request       contracts.CreateGatewayRequest
		wantErr       bool
		errorContains string
	}{
		{
			name:       "creates gateway",
			repository: &mockGatewayRepository{},
			request: contracts.CreateGatewayRequest{
				GatewayEUI: "0102030405060708",
				Latitude:   float32Pointer(37.5),
				Longitude:  float32Pointer(15),
			},
		},
		{
			name: "rejects duplicate GatewayEUI",
			repository: &mockGatewayRepository{
				gateways: []contracts.Gateway{{
					GatewayEUI: "0102030405060708",
				}},
			},
			request: contracts.CreateGatewayRequest{
				GatewayEUI: "0102030405060708",
			},
			wantErr:       true,
			errorContains: "already exists",
		},
		{
			name: "returns repository read errors",
			repository: &mockGatewayRepository{
				getAllErr: errors.New("database error"),
			},
			wantErr:       true,
			errorContains: "get gateways",
		},
		{
			name: "returns repository write errors",
			repository: &mockGatewayRepository{
				saveErr: errors.New("database error"),
			},
			request: contracts.CreateGatewayRequest{
				GatewayEUI: "0102030405060708",
				Latitude:   float32Pointer(37.5),
				Longitude:  float32Pointer(15),
			},
			wantErr:       true,
			errorContains: "save gateway",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGatewayService(tt.repository)
			_, err := service.CreateGateway(tt.request)

			if (err != nil) != tt.wantErr {
				t.Fatalf("CreateGateway() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errorContains) {
				t.Errorf("CreateGateway() error = %q, want substring %q", err, tt.errorContains)
			}
		})
	}
}

func float32Pointer(value float32) *float32 {
	return &value
}
