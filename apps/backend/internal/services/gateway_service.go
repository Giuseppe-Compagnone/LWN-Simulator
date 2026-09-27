package services

import (
	"errors"
	"fmt"
	"sync"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/google/uuid"
	"lwn-simulator-backend/internal/apperrors"
	"lwn-simulator-backend/internal/database"
)

type GatewayRepository interface {
	GetAll() ([]contracts.Gateway, error)
	GetByID(id string) (contracts.Gateway, error)
	Save(gateways []contracts.Gateway) error
	Update(gateway contracts.Gateway) error
	Delete(gateway contracts.Gateway) error
}

type GatewayService struct {
	repository GatewayRepository
	mu         sync.Mutex
}

func NewGatewayService(repository GatewayRepository) *GatewayService {
	return &GatewayService{repository: repository}
}

func (s *GatewayService) CreateGateway(req contracts.CreateGatewayRequest) (contracts.Gateway, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ValidateGatewayRequest(req); err != nil {
		return contracts.Gateway{}, err
	}

	gateways, err := s.repository.GetAll()
	if err != nil {
		return contracts.Gateway{}, fmt.Errorf("get gateways: %w", err)
	}

	for _, gateway := range gateways {
		if gateway.GatewayEUI == req.GatewayEUI {
			return contracts.Gateway{}, apperrors.Conflict(
				"gateway with gatewayEUI %q already exists",
				req.GatewayEUI,
			)
		}
		if gateway.MacAddress == req.MacAddress {
			return contracts.Gateway{}, apperrors.Conflict(
				"gateway with macAddress %q already exists",
				req.MacAddress,
			)
		}
	}

	gateway := contracts.Gateway{
		ID:          uuid.NewString(),
		Active:      true,
		Name:        req.Name,
		Type:        req.Type,
		MacAddress:  req.MacAddress,
		GatewayEUI:  req.GatewayEUI,
		KeepAlive:   req.KeepAlive,
		GatewayIPv4: req.GatewayIPv4,
		GatewayPort: req.GatewayPort,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Altitude:    req.Altitude,
	}

	gateways = append(gateways, gateway)
	if err := s.repository.Save(gateways); err != nil {
		return contracts.Gateway{}, fmt.Errorf("save gateway: %w", err)
	}

	return gateway, nil
}

func (s *GatewayService) GetGateway(req contracts.GetGatewayRequest) (contracts.GetGatewayResponse, error) {
	gateway, err := s.repository.GetByID(req.ID)
	if err != nil {
		return contracts.GetGatewayResponse{}, wrapGatewayRepositoryError("get gateway by id", err)
	}
	return contracts.GetGatewayResponse{Gateway: gateway}, nil
}

func (s *GatewayService) GetGateways(req contracts.GetGatewaysRequest) (contracts.GetGatewaysResponse, error) {
	gateways, err := s.repository.GetAll()
	if err != nil {
		return contracts.GetGatewaysResponse{}, fmt.Errorf("get all gateways: %w", err)
	}
	return contracts.GetGatewaysResponse{Gateways: gateways}, nil
}

func (s *GatewayService) UpdateGateway(req contracts.UpdateGatewayRequest) (contracts.UpdateGatewayResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ValidateGateway(req.Gateway); err != nil {
		return contracts.UpdateGatewayResponse{}, err
	}

	if _, err := s.repository.GetByID(req.ID); err != nil {
		return contracts.UpdateGatewayResponse{}, wrapGatewayRepositoryError("get gateway by id", err)
	}

	gateways, err := s.repository.GetAll()
	if err != nil {
		return contracts.UpdateGatewayResponse{}, fmt.Errorf("get gateways for update: %w", err)
	}

	for _, existing := range gateways {
		if existing.ID == req.ID {
			continue
		}
		if existing.GatewayEUI == req.Gateway.GatewayEUI {
			return contracts.UpdateGatewayResponse{}, apperrors.Conflict(
				"gateway with gatewayEUI %q already exists",
				req.Gateway.GatewayEUI,
			)
		}
		if existing.MacAddress == req.Gateway.MacAddress {
			return contracts.UpdateGatewayResponse{}, apperrors.Conflict(
				"gateway with macAddress %q already exists",
				req.Gateway.MacAddress,
			)
		}
	}

	gateway := req.Gateway
	gateway.ID = req.ID
	if err := s.repository.Update(gateway); err != nil {
		return contracts.UpdateGatewayResponse{}, wrapGatewayRepositoryError("update gateway", err)
	}

	return contracts.UpdateGatewayResponse{Gateway: gateway}, nil
}

func (s *GatewayService) DeleteGateway(req contracts.DeleteGatewayRequest) (contracts.DeleteGatewayResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	gateway, err := s.repository.GetByID(req.ID)
	if err != nil {
		return contracts.DeleteGatewayResponse{}, wrapGatewayRepositoryError("get gateway by id", err)
	}
	if err := s.repository.Delete(gateway); err != nil {
		return contracts.DeleteGatewayResponse{}, wrapGatewayRepositoryError("delete gateway", err)
	}
	return contracts.DeleteGatewayResponse{}, nil
}

func wrapGatewayRepositoryError(operation string, err error) error {
	if errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("%w: %s: %v", apperrors.ErrNotFound, operation, err)
	}
	return fmt.Errorf("%s: %w", operation, err)
}
