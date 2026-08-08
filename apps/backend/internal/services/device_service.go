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

// CreateDevice creates a new device after validating the request
// and checking that the DevEUI is not already in use.
func (s *DeviceService) CreateDevice(
	req contracts.CreateDeviceRequest,
) (contracts.Device, error) {

	if err := validateCreateDeviceRequest(req); err != nil {
		return contracts.Device{}, err
	}

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
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}

	devices = append(devices, device)

	if err := s.repository.Save(devices); err != nil {
		return contracts.Device{}, fmt.Errorf("save device: %w", err)
	}

	return device, nil
}

func validateCreateDeviceRequest(
	req contracts.CreateDeviceRequest,
) error {
	if req.DevEUI == "" {
		return fmt.Errorf("devEUI is required")
	}

	if req.Latitude < -90 || req.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}

	if req.Longitude < -180 || req.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	return nil
}
