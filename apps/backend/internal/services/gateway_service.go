package services

import (
	"fmt"
	"sync"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/google/uuid"
	"lwn-simulator-backend/internal/apperrors"
)

type GatewayRepository interface {
	GetAll() ([]contracts.Gateway, error)
	Save(gateways []contracts.Gateway) error
}

type GatewayService struct {
	repository GatewayRepository
	mu         sync.Mutex
}

func NewGatewayService(
	repository GatewayRepository,
) *GatewayService {
	return &GatewayService{
		repository: repository,
	}
}

func (s *GatewayService) CreateGateway(
	req contracts.CreateGatewayRequest,
) (contracts.Gateway, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	gateways, err := s.repository.GetAll()
	if err != nil {
		return contracts.Gateway{}, fmt.Errorf("get gateways: %w", err)
	}

	for _, gateway := range gateways {
		if gateway.GatewayEUI == req.GatewayEUI {
			return contracts.Gateway{}, apperrors.Conflict(
				"gateway with GatewayEUI %q already exists",
				req.GatewayEUI,
			)
		}
	}

	gateway := contracts.Gateway{
		ID:         uuid.NewString(),
		GatewayEUI: req.GatewayEUI,
		Latitude:   *req.Latitude,
		Longitude:  *req.Longitude,
	}

	gateways = append(gateways, gateway)

	if err := s.repository.Save(gateways); err != nil {
		return contracts.Gateway{}, fmt.Errorf("save gateway: %w", err)
	}

	return gateway, nil
}
