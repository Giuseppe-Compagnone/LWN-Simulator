package repositories

import (
	"lwn-simulator-backend/internal/database"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

type GatewayRepository struct {
	repository *database.JSONRepository[contracts.Gateway]
}

func NewGatewayRepository(dataDir string) *GatewayRepository {
	return &GatewayRepository{
		repository: (*database.JSONRepository[contracts.Gateway])(database.NewJSONRepository[contracts.Gateway](
			dataDir,
			"gateways.json",
		)),
	}
}

func (r *GatewayRepository) GetAll() ([]contracts.Gateway, error) {
	return r.repository.GetAll()
}

func (r *GatewayRepository) Save(gateways []contracts.Gateway) error {
	return r.repository.Save(gateways)
}
