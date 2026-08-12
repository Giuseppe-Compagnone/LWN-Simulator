package services

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

type mockDeviceRepository struct {
	devices []contracts.Device

	getAllErr  error
	getByIDErr error
	saveErr    error
	updateErr  error
	deleteErr  error

	getByIDCalls int
	saveCalls    int
	updateCalls  int
	deleteCalls  int

	lastSavedDevices  []contracts.Device
	lastUpdatedDevice contracts.Device
	lastDeletedDevice contracts.Device
}

const (
	testDeviceID1 = "550e8400-e29b-41d4-a716-446655440000"
	testDeviceID2 = "6ba7b810-9dad-41d1-80b4-00c04fd430c8"
)

func (m *mockDeviceRepository) GetAll() ([]contracts.Device, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}

	return m.devices, nil
}

func (m *mockDeviceRepository) GetByID(id string) (contracts.Device, error) {
	m.getByIDCalls++

	if m.getByIDErr != nil {
		return contracts.Device{}, m.getByIDErr
	}

	for _, device := range m.devices {
		if device.ID == id {
			return device, nil
		}
	}

	return contracts.Device{}, errors.New("device not found")
}

func (m *mockDeviceRepository) Save(devices []contracts.Device) error {
	m.saveCalls++

	if m.saveErr != nil {
		return m.saveErr
	}

	m.devices = devices
	m.lastSavedDevices = devices

	return nil
}

func (m *mockDeviceRepository) Update(device contracts.Device) error {
	m.updateCalls++

	if m.updateErr != nil {
		return m.updateErr
	}

	m.lastUpdatedDevice = device

	for i, existing := range m.devices {
		if existing.ID == device.ID {
			m.devices[i] = device
			return nil
		}
	}

	return errors.New("device not found")
}

func (m *mockDeviceRepository) Delete(device contracts.Device) error {
	m.deleteCalls++

	if m.deleteErr != nil {
		return m.deleteErr
	}

	m.lastDeletedDevice = device

	for i, existing := range m.devices {
		if existing.ID == device.ID {
			m.devices = append(m.devices[:i], m.devices[i+1:]...)
			return nil
		}
	}

	return errors.New("device not found")
}

func newTestDeviceService(
	repository *mockDeviceRepository,
) *DeviceService {
	return NewDeviceService(repository)
}

func TestDeviceService_CreateDevice(t *testing.T) {
	tests := []struct {
		name          string
		repository    *mockDeviceRepository
		request       contracts.CreateDeviceRequest
		wantErr       bool
		errorContains string
		wantDevEUI    string
		wantLatitude  float32
		wantLongitude float32
	}{
		{
			name: "creates device",
			repository: &mockDeviceRepository{
				devices: []contracts.Device{},
			},
			request: contracts.CreateDeviceRequest{
				DevEUI:    "0102030405060708",
				Latitude:  float32Ptr(37.5),
				Longitude: float32Ptr(15.0),
			},
			wantDevEUI:    "0102030405060708",
			wantLatitude:  37.5,
			wantLongitude: 15.0,
		},
		{
			name: "returns error when DevEUI already exists",
			repository: &mockDeviceRepository{
				devices: []contracts.Device{
					{
						ID:     testDeviceID1,
						DevEUI: "0102030405060708",
					},
				},
			},
			request: contracts.CreateDeviceRequest{
				DevEUI:    "0102030405060708",
				Latitude:  float32Ptr(37.5),
				Longitude: float32Ptr(15.0),
			},
			wantErr:       true,
			errorContains: "already exists",
		},
		{
			name: "returns error when GetAll fails",
			repository: &mockDeviceRepository{
				getAllErr: errors.New("database error"),
			},
			request: contracts.CreateDeviceRequest{
				DevEUI:    "0102030405060708",
				Latitude:  float32Ptr(37.5),
				Longitude: float32Ptr(15.0),
			},
			wantErr:       true,
			errorContains: "get devices",
		},
		{
			name: "returns error when Save fails",
			repository: &mockDeviceRepository{
				devices: []contracts.Device{},
				saveErr: errors.New("database error"),
			},
			request: contracts.CreateDeviceRequest{
				DevEUI:    "0102030405060708",
				Latitude:  float32Ptr(37.5),
				Longitude: float32Ptr(15.0),
			},
			wantErr:       true,
			errorContains: "save device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newTestDeviceService(tt.repository)

			got, err := service.CreateDevice(tt.request)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"CreateDevice() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf(
						"CreateDevice() error = %q, want substring %q",
						err,
						tt.errorContains,
					)
				}

				return
			}

			if got.Device.ID == "" {
				t.Error("CreateDevice() generated an empty ID")
			}

			if got.Device.DevEUI != tt.wantDevEUI {
				t.Errorf(
					"DevEUI = %q, want %q",
					got.Device.DevEUI,
					tt.wantDevEUI,
				)
			}

			if got.Device.Latitude != tt.wantLatitude {
				t.Errorf(
					"Latitude = %v, want %v",
					got.Device.Latitude,
					tt.wantLatitude,
				)
			}

			if got.Device.Longitude != tt.wantLongitude {
				t.Errorf(
					"Longitude = %v, want %v",
					got.Device.Longitude,
					tt.wantLongitude,
				)
			}

			if tt.repository.saveCalls != 1 {
				t.Errorf(
					"Save() calls = %d, want 1",
					tt.repository.saveCalls,
				)
			}
		})
	}
}

func float32Ptr(value float32) *float32 {
	return &value
}

func TestDeviceService_GetDevice(t *testing.T) {
	tests := []struct {
		name          string
		repository    *mockDeviceRepository
		request       contracts.GetDeviceRequest
		want          contracts.Device
		wantErr       bool
		errorContains string
	}{
		{
			name: "returns device",
			repository: &mockDeviceRepository{
				devices: []contracts.Device{
					{
						ID:        testDeviceID1,
						DevEUI:    "0102030405060708",
						Latitude:  37.5,
						Longitude: 15.0,
					},
				},
			},
			request: contracts.GetDeviceRequest{
				ID: testDeviceID1,
			},
			want: contracts.Device{
				ID:        testDeviceID1,
				DevEUI:    "0102030405060708",
				Latitude:  37.5,
				Longitude: 15.0,
			},
		},
		{
			name: "returns error when repository fails",
			repository: &mockDeviceRepository{
				getByIDErr: errors.New("database error"),
			},
			request: contracts.GetDeviceRequest{
				ID: testDeviceID1,
			},
			wantErr:       true,
			errorContains: "get device by id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newTestDeviceService(tt.repository)

			got, err := service.GetDevice(tt.request)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"GetDevice() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf(
						"GetDevice() error = %q, want substring %q",
						err,
						tt.errorContains,
					)
				}

				return
			}

			if !reflect.DeepEqual(got.Device, tt.want) {
				t.Errorf(
					"GetDevice() = %v, want %v",
					got.Device,
					tt.want,
				)
			}
		})
	}
}

func TestDeviceService_GetDevices(t *testing.T) {
	tests := []struct {
		name          string
		repository    *mockDeviceRepository
		want          []contracts.Device
		wantErr       bool
		errorContains string
	}{
		{
			name: "returns all devices",
			repository: &mockDeviceRepository{
				devices: []contracts.Device{
					{
						ID:     testDeviceID1,
						DevEUI: "0102030405060708",
					},
					{
						ID:     testDeviceID2,
						DevEUI: "1122334455667788",
					},
				},
			},
			want: []contracts.Device{
				{
					ID:     testDeviceID1,
					DevEUI: "0102030405060708",
				},
				{
					ID:     testDeviceID2,
					DevEUI: "1122334455667788",
				},
			},
		},
		{
			name: "returns error when repository fails",
			repository: &mockDeviceRepository{
				getAllErr: errors.New("database error"),
			},
			wantErr:       true,
			errorContains: "get all devices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newTestDeviceService(tt.repository)

			got, err := service.GetDevices(
				contracts.GetDevicesRequest{},
			)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"GetDevices() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf(
						"GetDevices() error = %q, want substring %q",
						err,
						tt.errorContains,
					)
				}

				return
			}

			if got.Devices == nil {
				t.Fatal("GetDevices() returned nil Devices")
			}

			if !reflect.DeepEqual(got.Devices, tt.want) {
				t.Errorf(
					"GetDevices() = %v, want %v",
					got.Devices,
					tt.want,
				)
			}
		})
	}
}

func TestDeviceService_UpdateDevice(t *testing.T) {
	tests := []struct {
		name          string
		repository    *mockDeviceRepository
		request       contracts.UpdateDeviceRequest
		want          contracts.Device
		wantErr       bool
		errorContains string
	}{
		{
			name: "updates DevEUI",
			repository: &mockDeviceRepository{
				devices: []contracts.Device{
					{
						ID:        testDeviceID1,
						DevEUI:    "0102030405060708",
						Latitude:  37.5,
						Longitude: 15.0,
					},
				},
			},
			request: contracts.UpdateDeviceRequest{
				ID: testDeviceID1,
				DevEUI: stringPtr(
					"1122334455667788",
				),
			},
			want: contracts.Device{
				ID:        testDeviceID1,
				DevEUI:    "1122334455667788",
				Latitude:  37.5,
				Longitude: 15.0,
			},
		},
		{
			name: "updates latitude",
			repository: &mockDeviceRepository{
				devices: []contracts.Device{
					{
						ID:        testDeviceID1,
						DevEUI:    "0102030405060708",
						Latitude:  37.5,
						Longitude: 15.0,
					},
				},
			},
			request: contracts.UpdateDeviceRequest{
				ID:       testDeviceID1,
				Latitude: float32Ptr(40.0),
			},
			want: contracts.Device{
				ID:        testDeviceID1,
				DevEUI:    "0102030405060708",
				Latitude:  40.0,
				Longitude: 15.0,
			},
		},
		{
			name: "updates longitude",
			repository: &mockDeviceRepository{
				devices: []contracts.Device{
					{
						ID:        testDeviceID1,
						DevEUI:    "0102030405060708",
						Latitude:  37.5,
						Longitude: 15.0,
					},
				},
			},
			request: contracts.UpdateDeviceRequest{
				ID:        testDeviceID1,
				Longitude: float32Ptr(20.0),
			},
			want: contracts.Device{
				ID:        testDeviceID1,
				DevEUI:    "0102030405060708",
				Latitude:  37.5,
				Longitude: 20.0,
			},
		},
		{
			name: "updates all fields",
			repository: &mockDeviceRepository{
				devices: []contracts.Device{
					{
						ID:        testDeviceID1,
						DevEUI:    "0102030405060708",
						Latitude:  37.5,
						Longitude: 15.0,
					},
				},
			},
			request: contracts.UpdateDeviceRequest{
				ID:        testDeviceID1,
				DevEUI:    stringPtr("1122334455667788"),
				Latitude:  float32Ptr(40.0),
				Longitude: float32Ptr(20.0),
			},
			want: contracts.Device{
				ID:        testDeviceID1,
				DevEUI:    "1122334455667788",
				Latitude:  40.0,
				Longitude: 20.0,
			},
		},
		{
			name: "returns error when device does not exist",
			repository: &mockDeviceRepository{
				devices: []contracts.Device{},
			},
			request: contracts.UpdateDeviceRequest{
				ID:       "unknown-device",
				Latitude: float32Ptr(40.0),
			},
			wantErr:       true,
			errorContains: "get device by id",
		},
		{
			name: "returns error when update fails",
			repository: &mockDeviceRepository{
				devices: []contracts.Device{
					{
						ID: testDeviceID1,
					},
				},
				updateErr: errors.New("database error"),
			},
			request: contracts.UpdateDeviceRequest{
				ID:       testDeviceID1,
				Latitude: float32Ptr(40.0),
			},
			wantErr:       true,
			errorContains: "update device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newTestDeviceService(tt.repository)

			got, err := service.UpdateDevice(tt.request)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"UpdateDevice() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf(
						"UpdateDevice() error = %q, want substring %q",
						err,
						tt.errorContains,
					)
				}

				return
			}

			if !reflect.DeepEqual(got.Device, tt.want) {
				t.Errorf(
					"UpdateDevice() = %v, want %v",
					got.Device,
					tt.want,
				)
			}

			if !reflect.DeepEqual(
				tt.repository.lastUpdatedDevice,
				tt.want,
			) {
				t.Errorf(
					"Update() received = %v, want %v",
					tt.repository.lastUpdatedDevice,
					tt.want,
				)
			}
		})
	}
}

func stringPtr(value string) *string {
	return &value
}

func TestDeviceService_DeleteDevice(t *testing.T) {
	tests := []struct {
		name          string
		repository    *mockDeviceRepository
		request       contracts.DeleteDeviceRequest
		wantErr       bool
		errorContains string
	}{
		{
			name: "deletes device",
			repository: &mockDeviceRepository{
				devices: []contracts.Device{
					{
						ID:     testDeviceID1,
						DevEUI: "0102030405060708",
					},
				},
			},
			request: contracts.DeleteDeviceRequest{
				ID: testDeviceID1,
			},
		},
		{
			name: "returns error when device does not exist",
			repository: &mockDeviceRepository{
				devices: []contracts.Device{},
			},
			request: contracts.DeleteDeviceRequest{
				ID: "unknown-device",
			},
			wantErr:       true,
			errorContains: "get device by id",
		},
		{
			name: "returns error when delete fails",
			repository: &mockDeviceRepository{
				devices: []contracts.Device{
					{
						ID: testDeviceID1,
					},
				},
				deleteErr: errors.New("database error"),
			},
			request: contracts.DeleteDeviceRequest{
				ID: testDeviceID1,
			},
			wantErr:       true,
			errorContains: "delete device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newTestDeviceService(tt.repository)

			_, err := service.DeleteDevice(tt.request)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"DeleteDevice() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf(
						"DeleteDevice() error = %q, want substring %q",
						err,
						tt.errorContains,
					)
				}

				return
			}

			if tt.repository.deleteCalls != 1 {
				t.Errorf(
					"Delete() calls = %d, want 1",
					tt.repository.deleteCalls,
				)
			}

			if tt.repository.lastDeletedDevice.ID != testDeviceID1 {
				t.Errorf(
					"Delete() received device = %v",
					tt.repository.lastDeletedDevice,
				)
			}

			if len(tt.repository.devices) != 0 {
				t.Errorf(
					"devices after DeleteDevice() = %v, want empty",
					tt.repository.devices,
				)
			}
		})
	}
}
