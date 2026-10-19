package runtime

import (
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/lorawan"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

func TestRealGatewayPayloadIsDecodedAfterDecryption(t *testing.T) {
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000210")
	configureABP(&device)
	device.PayloadConfig.PayloadCodec = &contracts.PayloadCodec{
		Enabled: true,
		Script:  "function decodeUplink(input) { return { data: { first: input.bytes[0] } }; }",
	}
	session := newDeviceSession(device)
	port := byte(device.FrameConfig.FPort)
	frame, err := lorawan.BuildDataFrame(lorawan.DataFrameOptions{
		DevAddr:   0x26011bda,
		FCnt:      1,
		FPort:     &port,
		Payload:   []byte{42},
		NwkSKey:   session.NwkSKey,
		AppSKey:   session.AppSKey,
		Direction: 0,
	})
	if err != nil {
		t.Fatalf("build uplink: %v", err)
	}
	event := contracts.SimulationEvent{}
	decorateDecodedRadioPayload(&event, device, &types.RadioTransmission{
		RadioFrame: types.RadioFrame{Payload: frame},
	}, session)
	if event.DecodedPayload == nil || (*event.DecodedPayload)["first"] != float64(42) {
		t.Fatalf("unexpected decoded real-gateway payload: %+v", event.DecodedPayload)
	}
}
