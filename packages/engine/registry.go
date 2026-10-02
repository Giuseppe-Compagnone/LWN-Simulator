package engine

import (
	"sort"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

type Registry struct {
	devices  map[string]contracts.Device
	gateways map[string]contracts.Gateway
}

func NewRegistry(devices []contracts.Device, gateways []contracts.Gateway) (*Registry, error) {
	if err := ValidateHardware(devices, gateways); err != nil {
		return nil, err
	}

	registry := &Registry{
		devices:  make(map[string]contracts.Device, len(devices)),
		gateways: make(map[string]contracts.Gateway, len(gateways)),
	}

	for _, device := range devices {
		registry.devices[device.ID] = device
	}
	for _, gateway := range gateways {
		registry.gateways[gateway.ID] = gateway
	}

	return registry, nil
}

func (r *Registry) Device(id string) (contracts.Device, bool) {
	device, ok := r.devices[id]
	return device, ok
}

func (r *Registry) Gateway(id string) (contracts.Gateway, bool) {
	gateway, ok := r.gateways[id]
	return gateway, ok
}

func (r *Registry) Devices() []contracts.Device {
	ids := make([]string, 0, len(r.devices))
	for id := range r.devices {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	devices := make([]contracts.Device, 0, len(ids))
	for _, id := range ids {
		devices = append(devices, r.devices[id])
	}
	return devices
}

func (r *Registry) Gateways() []contracts.Gateway {
	ids := make([]string, 0, len(r.gateways))
	for id := range r.gateways {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	gateways := make([]contracts.Gateway, 0, len(ids))
	for _, id := range ids {
		gateways = append(gateways, r.gateways[id])
	}
	return gateways
}

func (r *Registry) Snapshot() contracts.SimulationSnapshot {
	devices := r.Devices()
	runtimeDevices := make([]contracts.RuntimeDevice, 0, len(devices))
	for _, device := range devices {
		status := contracts.RuntimeStatusRegistered
		if device.Active {
			status = contracts.RuntimeStatusActive
		}
		runtimeDevices = append(runtimeDevices, contracts.RuntimeDevice{
			ID:     device.ID,
			Name:   device.Name,
			Active: device.Active,
			Status: status,
		})
	}

	gateways := r.Gateways()
	runtimeGateways := make([]contracts.RuntimeGateway, 0, len(gateways))
	for _, gateway := range gateways {
		status := contracts.RuntimeStatusRegistered
		if gateway.Active {
			status = contracts.RuntimeStatusActive
		}
		runtimeGateways = append(runtimeGateways, contracts.RuntimeGateway{
			ID:     gateway.ID,
			Name:   gateway.Name,
			Active: gateway.Active,
			Status: status,
		})
	}

	return contracts.SimulationSnapshot{
		Devices:  runtimeDevices,
		Gateways: runtimeGateways,
	}
}
