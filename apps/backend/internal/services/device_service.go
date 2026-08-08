package services

import (
	"fmt"
	"lwn-simulator-backend/internal/repositories"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/google/uuid"
)

type DeviceService struct {
	repository *repositories.DeviceRepository
}

func NewDeviceService(
	repository *repositories.DeviceRepository,
) *DeviceService {
	return &DeviceService{
		repository: repository,
	}
}

func (s *DeviceService) CreateDevice(
	req contracts.CreateDeviceRequest,
) (contracts.Device, error) {

	devices, err := s.repository.GetAll()
	if err != nil {
		return contracts.Device{}, fmt.Errorf("get devices: %w", err)
	}

	for _, device := range devices {
		if device.DevEUI == req.DevEUI {
			return contracts.Device{}, fmt.Errorf(
				"device with DevEUI %q already exists",
				req.DevEUI,
			)
		}
	}

	device := contracts.Device{
		ID:        uuid.NewString(),
		DevEUI:    req.DevEUI,
		Latitude:  *req.Latitude,
		Longitude: *req.Longitude,
	}

	devices = append(devices, device)

	if err := s.repository.Save(devices); err != nil {
		return contracts.Device{}, fmt.Errorf("save device: %w", err)
	}

	return device, nil
}
