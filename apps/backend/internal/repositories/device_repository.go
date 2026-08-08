package repositories

import (
	"lwn-simulator-backend/internal/database"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

type DeviceRepository struct {
	repository *database.JSONRepository[contracts.Device]
}

func NewDeviceRepository(dataDir string) *DeviceRepository {
	return &DeviceRepository{
		repository: database.NewJSONRepository[contracts.Device](
			dataDir,
			"devices.json",
		),
	}
}

func (r *DeviceRepository) GetAll() ([]contracts.Device, error) {
	return r.repository.GetAll()
}

func (r *DeviceRepository) Save(devices []contracts.Device) error {
	return r.repository.Save(devices)
}
