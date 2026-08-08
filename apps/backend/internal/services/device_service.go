package services

import (
	"fmt"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/google/uuid"
)

type DeviceRepository interface {
	GetAll() ([]contracts.Device, error)
	GetByID(id string) (contracts.Device, error)
	Save(devices []contracts.Device) error
	Update(device contracts.Device) error
	Delete(device contracts.Device) error
}
type DeviceService struct {
	repository DeviceRepository
}

func NewDeviceService(
	repository DeviceRepository,
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

func (s *DeviceService) GetDevices(
	req contracts.GetDevicesRequest,
) (contracts.GetDevicesResponse, error) {
	devices, err := s.repository.GetAll()

	if err != nil {
		return contracts.GetDevicesResponse{}, fmt.Errorf("get all devices: %w", err)
	}

	return contracts.GetDevicesResponse{
		Devices: &devices,
	}, nil
}

func (s *DeviceService) UpdateDevice(
	req contracts.UpdateDeviceRequest,
) (contracts.UpdateDeviceResponse, error) {
	device, err := s.repository.GetByID(req.ID)

	if err != nil {
		return contracts.UpdateDeviceResponse{}, fmt.Errorf("get device by id: %w", err)
	}

	if req.DevEUI != nil {
		device.DevEUI = *req.DevEUI
	}

	if req.Latitude != nil {
		device.Latitude = *req.Latitude
	}

	if req.Longitude != nil {
		device.Longitude = *req.Longitude
	}

	if err := s.repository.Update(device); err != nil {
		return contracts.UpdateDeviceResponse{}, fmt.Errorf(
			"update device: %w",
			err,
		)
	}

	return contracts.UpdateDeviceResponse{
		Device: device,
	}, nil
}

func (s *DeviceService) DeleteDevice(
	req contracts.DeleteDeviceRequest,
) (contracts.DeleteDeviceResponse, error) {
	device, err := s.repository.GetByID(req.ID)

	if err != nil {
		return contracts.DeleteDeviceResponse{}, fmt.Errorf("get device by id: %w", err)
	}

	if err := s.repository.Delete(device); err != nil {
		return contracts.DeleteDeviceResponse{}, fmt.Errorf(
			"delete device: %w",
			err,
		)
	}

	return contracts.DeleteDeviceResponse{}, nil
}
