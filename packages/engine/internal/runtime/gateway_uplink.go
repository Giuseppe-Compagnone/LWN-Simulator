package runtime

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/lorawan"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

// buildUplinkPayloadLocked creates the PHYPayload that a virtual gateway
// exposes to an external Gateway Bridge. The simulator keeps its radio model
// independent from the transport, so this conversion is done only at the
// gateway boundary.
func (e *Engine) buildUplinkPayloadLocked(
	device contracts.Device,
	session *types.DeviceSession,
	frameCounter int64,
	confirmed bool,
	fragmentIndex int,
	dataRate int,
) []byte {
	if session == nil {
		return nil
	}
	if session.PendingJoinRequest != nil && session.PendingJoinRequest.JoinEUI != "" {
		if device.OOTAConfig == nil {
			return nil
		}
		joinEUI, ok := parseEUI64(session.PendingJoinRequest.JoinEUI)
		if !ok {
			return nil
		}
		devEUI, ok := parseEUI64(device.DevEUI)
		if !ok {
			return nil
		}
		appKey, err := lorawan.DecodeHex(device.OOTAConfig.AppKey)
		if err != nil {
			return nil
		}
		payload, err := lorawan.BuildJoinRequest(lorawan.JoinRequestOptions{
			JoinEUI:  joinEUI,
			DevEUI:   devEUI,
			DevNonce: session.DevNonce,
			AppKey:   appKey,
		})
		if err != nil {
			return nil
		}
		return payload
	}

	devAddrBytes, err := hex.DecodeString(session.DeviceAddress)
	if err != nil || len(devAddrBytes) != 4 || len(session.NwkSKey) != 16 || len(session.AppSKey) != 16 {
		return nil
	}
	var fPort *byte
	if device.FrameConfig.FPort > 0 {
		port := byte(device.FrameConfig.FPort)
		fPort = &port
	}
	payload := payloadBytes(device)
	maximum := maximumPayloadSize(device.LocationConfig.Region, dataRate)
	if maximum > 0 && len(payload) > maximum {
		start := fragmentIndex * maximum
		if start >= len(payload) {
			payload = nil
		} else {
			end := start + maximum
			if end > len(payload) {
				end = len(payload)
			}
			payload = payload[start:end]
		}
	}
	devAddr := binary.BigEndian.Uint32(devAddrBytes)
	frame, err := lorawan.BuildDataFrame(lorawan.DataFrameOptions{
		DevAddr:   devAddr,
		FCnt:      uint32(maxInt64(frameCounter, 0)),
		FPort:     fPort,
		Payload:   payload,
		Confirmed: confirmed,
		ADR:       device.AdvancedConfig.ADREnabled,
		Direction: 0,
		NwkSKey:   session.NwkSKey,
		AppSKey:   session.AppSKey,
	})
	if err != nil {
		return nil
	}
	return frame
}

func parseEUI64(value string) (uint64, bool) {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 8 {
		return 0, false
	}
	return binary.BigEndian.Uint64(decoded), true
}

func maxInt64(value, minimum int64) int64 {
	if value < minimum {
		return minimum
	}
	return value
}

func dataRateString(spreadingFactor int, bandwidth int64) string {
	if spreadingFactor < 7 || spreadingFactor > 12 || bandwidth <= 0 {
		return ""
	}
	return fmt.Sprintf("SF%dBW%d", spreadingFactor, bandwidth/1000)
}
