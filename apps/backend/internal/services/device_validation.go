package services

import (
	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine"
)

const validationDeviceID = "00000000-0000-4000-8000-000000000000"

func ValidateDeviceRequest(req contracts.CreateDeviceRequest) error {
	return ValidateDevice(contracts.Device{
		ID: validationDeviceID, Active: true, Name: req.Name, DevEUI: req.DevEUI,
		Class: req.Class, Activation: req.Activation, OOTAConfig: req.OOTAConfig,
		ABPConfig: req.ABPConfig, LocationConfig: req.LocationConfig,
		RX1Config: req.RX1Config, RX2Config: req.RX2Config,
		FrameConfig: req.FrameConfig, PayloadConfig: req.PayloadConfig,
		AdvancedConfig: req.AdvancedConfig,
	})
}

func ValidateDevice(device contracts.Device) error {
	return engine.ValidateHardware([]contracts.Device{device}, nil)
}
