package services

import (
	"fmt"
	"lwn-simulator-backend/internal/repositories"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type GatewayService struct {
	repository *repositories.GatewayRepository
	validator  *validator.Validate
}

func NewGatewayService(
	repository *repositories.GatewayRepository,
	validator *validator.Validate,
) *GatewayService {
	return &GatewayService{
		repository: repository,
		validator:  validator,
	}
}

func (s *GatewayService) CreateGateway(
	req contracts.CreateGatewayRequest,
) (contracts.Gateway, error) {

	gateways, err := s.repository.GetAll()
	if err != nil {
		return contracts.Gateway{}, fmt.Errorf("get gateways: %w", err)
	}

	for _, gateway := range gateways {
		if gateway.GatewayEUI == req.GatewayEUI {
			return contracts.Gateway{}, fmt.Errorf(
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
