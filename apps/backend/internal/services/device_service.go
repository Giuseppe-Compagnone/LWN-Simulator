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
) (contracts.CreateDeviceResponse, error) {

	devices, err := s.repository.GetAll()
	if err != nil {
		return contracts.CreateDeviceResponse{}, fmt.Errorf("get devices: %w", err)
	}

	for _, device := range devices {
		if device.DevEUI == req.DevEUI {
			return contracts.CreateDeviceResponse{}, fmt.Errorf(
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
		return contracts.CreateDeviceResponse{}, fmt.Errorf("save device: %w", err)
	}

	return contracts.CreateDeviceResponse{
		Device: device,
	}, nil
}

func (s *DeviceService) GetDevice(
	req contracts.GetDeviceRequest,
) (contracts.GetDeviceResponse, error) {

	device, err := s.repository.GetByID(req.ID)

	if err != nil {
		return contracts.GetDeviceResponse{}, fmt.Errorf("get device by id: %w", err)
	}

	return contracts.GetDeviceResponse{
		Device: device,
	}, nil
}
