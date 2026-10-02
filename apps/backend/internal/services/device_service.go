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

type DeviceRepository interface {
	GetAll() ([]contracts.Device, error)
	GetByID(id string) (contracts.Device, error)
	Save(devices []contracts.Device) error
	Update(device contracts.Device) error
	Delete(device contracts.Device) error
}
type DeviceService struct {
	repository DeviceRepository
	mu         sync.Mutex
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
	s.mu.Lock()
	defer s.mu.Unlock()

	devices, err := s.repository.GetAll()
	if err != nil {
		return contracts.CreateDeviceResponse{}, fmt.Errorf("get devices: %w", err)
	}

	for _, device := range devices {
		if device.DevEUI == req.DevEUI {
			return contracts.CreateDeviceResponse{}, apperrors.Conflict(
				"device with DevEUI %q already exists",
				req.DevEUI,
			)
		}
	}

	device := contracts.Device{
		ID:             uuid.NewString(),
		Active:         true,
		DevEUI:         req.DevEUI,
		Name:           req.Name,
		Activation:     req.Activation,
		Class:          req.Class,
		LocationConfig: req.LocationConfig,
		ABPConfig:      req.ABPConfig,
		OOTAConfig:     req.OOTAConfig,
		RX1Config:      req.RX1Config,
		RX2Config:      req.RX2Config,
		AdvancedConfig: req.AdvancedConfig,
		FrameConfig:    req.FrameConfig,
		PayloadConfig:  req.PayloadConfig,
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
		return contracts.GetDeviceResponse{}, wrapDeviceRepositoryError(
			"get device by id",
			err,
		)
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
		Devices: devices,
	}, nil
}

func (s *DeviceService) UpdateDevice(
	req contracts.UpdateDeviceRequest,
) (contracts.UpdateDeviceResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.repository.GetByID(req.ID)
	if err != nil {
		return contracts.UpdateDeviceResponse{}, wrapDeviceRepositoryError(
			"get device by id",
			err,
		)
	}

	devices, err := s.repository.GetAll()
	if err != nil {
		return contracts.UpdateDeviceResponse{}, fmt.Errorf(
			"get devices for update: %w",
			err,
		)
	}

	for _, existing := range devices {
		if existing.ID != req.ID && existing.DevEUI == req.Device.DevEUI {
			return contracts.UpdateDeviceResponse{}, apperrors.Conflict(
				"device with DevEUI %q already exists",
				req.Device.DevEUI,
			)
		}
	}

	device := req.Device
	device.ID = req.ID

	if err := s.repository.Update(device); err != nil {
		return contracts.UpdateDeviceResponse{}, wrapDeviceRepositoryError(
			"update device",
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
	s.mu.Lock()
	defer s.mu.Unlock()
	device, err := s.repository.GetByID(req.ID)

	if err != nil {
		return contracts.DeleteDeviceResponse{}, wrapDeviceRepositoryError(
			"get device by id",
			err,
		)
	}

	if err := s.repository.Delete(device); err != nil {
		return contracts.DeleteDeviceResponse{}, wrapDeviceRepositoryError(
			"delete device",
			err,
		)
	}

	return contracts.DeleteDeviceResponse{}, nil
}

func wrapDeviceRepositoryError(operation string, err error) error {
	if errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("%w: %s: %v", apperrors.ErrNotFound, operation, err)
	}

	return fmt.Errorf("%s: %w", operation, err)
}
