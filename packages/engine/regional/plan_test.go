package regional

import (
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

func TestEveryContractRegionHasAUsablePlan(t *testing.T) {
	regions := []contracts.DeviceRegion{
		contracts.EU868, contracts.US915, contracts.CN779, contracts.EU433, contracts.AU915,
		contracts.CN470, contracts.AS923, contracts.KR920, contracts.IN865, contracts.RU864,
	}
	for _, region := range regions {
		plan, ok := Plan(region)
		if !ok {
			t.Fatalf("missing plan for %s", region)
		}
		if plan.RX2Frequency <= 0 || plan.BeaconFrequency <= 0 || plan.PingSlotFrequency <= 0 || len(UplinkChannels(region, plan.DefaultDataRate)) == 0 {
			t.Fatalf("incomplete plan for %s: %+v", region, plan)
		}
	}
}

func TestRU864UsesDistinctBeaconAndPingSlotFrequencies(t *testing.T) {
	plan := MustPlan(contracts.RU864)
	if plan.BeaconFrequency != 869_100_000 || plan.PingSlotFrequency != 868_900_000 {
		t.Fatalf("unexpected RU864 Class B plan: %+v", plan)
	}
}

func TestRegionalChannelPlansRemainDistinct(t *testing.T) {
	if got := UplinkChannels(contracts.CN470, 5); len(got) != 96 || got[0].Frequency != 470_300_000 || got[95].Frequency != 489_300_000 {
		t.Fatalf("unexpected CN470 channels: first=%+v count=%d last=%+v", got[0], len(got), got[len(got)-1])
	}
	if got := UplinkChannels(contracts.CN779, 5); len(got) != 3 || got[0].Frequency != 779_500_000 {
		t.Fatalf("unexpected CN779 channels: %+v", got)
	}
	if got := UplinkChannels(contracts.US915, 4); len(got) != 8 || got[0].Bandwidth != 500_000 {
		t.Fatalf("unexpected US915 DR4 channels: %+v", got)
	}
	if got := UplinkChannels(contracts.AU915, 5); len(got) != 64 || got[63].Frequency != 927_800_000 {
		t.Fatalf("unexpected AU915 channels: count=%d last=%+v", len(got), got[len(got)-1])
	}
}

func TestDataRateProfilesUseRegionalModulationAndPayloadLimits(t *testing.T) {
	if profile := Profile(contracts.US915, 0); profile.SpreadingFactor != 10 || profile.MaximumPayload != 11 {
		t.Fatalf("unexpected US915 DR0 profile: %+v", profile)
	}
	if profile := Profile(contracts.EU868, 2); profile.SpreadingFactor != 10 || profile.MaximumPayload != 51 {
		t.Fatalf("unexpected EU868 DR2 profile: %+v", profile)
	}
	if profile := Profile(contracts.AU915, 6); profile.Bandwidth != 500_000 || profile.MaximumPayload != 222 {
		t.Fatalf("unexpected AU915 DR6 profile: %+v", profile)
	}
}

func TestDataRateForModulationIsRegionAware(t *testing.T) {
	if dataRate, ok := DataRateForModulation(contracts.US915, 10, 125_000); !ok || dataRate != 0 {
		t.Fatalf("unexpected US915 SF10BW125 mapping: DR%d, %v", dataRate, ok)
	}
	if dataRate, ok := DataRateForModulation(contracts.EU868, 10, 125_000); !ok || dataRate != 2 {
		t.Fatalf("unexpected EU868 SF10BW125 mapping: DR%d, %v", dataRate, ok)
	}
	if _, ok := DataRateForModulation(contracts.EU868, 6, 125_000); ok {
		t.Fatal("unexpected mapping for unsupported modulation")
	}
}
