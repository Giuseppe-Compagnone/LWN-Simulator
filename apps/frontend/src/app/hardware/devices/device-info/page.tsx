"use client";

import { Device } from "@lwn-simulator/contracts";
import { useDeviceService } from "@lwn-simulator/sdk";
import {
  Button,
  EntityDetailTone,
  EntityDetailSection,
  EntityDetails,
  EntityDetailValueFormat,
  EntityDetailsSectionLayout,
  PageHeader,
  Spinner,
} from "@lwn-simulator/ui-components";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";

const displayValue = (value: unknown) =>
  value === undefined || value === null || value === ""
    ? "Not configured"
    : String(value);

const displayBoolean = (value: boolean) => (value ? "Enabled" : "Disabled");

const getDeviceSections = (device: Device): Array<EntityDetailSection> => {
  const sections: Array<EntityDetailSection> = [
    {
      title: "Overview",
      icon: "memory",
      description: "Identity and current runtime state",
      items: [
        { label: "ID", value: device.id, mono: true },
        {
          label: "Status",
          value: device.active ? "Active" : "Inactive",
          tone: device.active
            ? EntityDetailTone.Positive
            : EntityDetailTone.Muted,
        },
        { label: "Name", value: device.name },
        { label: "DevEUI", value: device.devEUI, mono: true },
        {
          label: "Class",
          value: device.class,
          format: EntityDetailValueFormat.Enum,
        },
        {
          label: "Activation",
          value: device.activation,
          format: EntityDetailValueFormat.Enum,
        },
      ],
    },
    {
      title: "Location",
      icon: "location_on",
      items: [
        { label: "Latitude", value: device.locationConfig.latitude },
        { label: "Longitude", value: device.locationConfig.longitude },
        { label: "Altitude", value: device.locationConfig.altitude },
        {
          label: "Region",
          value: device.locationConfig.region,
          format: EntityDetailValueFormat.Enum,
        },
      ],
    },
    {
      title: "RX1 Configuration",
      icon: "settings_input_antenna",
      items: [
        { label: "Delay", value: device.RX1Config.delay },
        { label: "Duration", value: device.RX1Config.duration },
        {
          label: "Data rate offset",
          value: device.RX1Config.dataRateOffset,
        },
      ],
    },
    {
      title: "RX2 Configuration",
      icon: "settings_input_component",
      items: [
        { label: "Delay", value: device.RX2Config.delay },
        { label: "Duration", value: device.RX2Config.duration },
        {
          label: "Channel frequency",
          value: device.RX2Config.channelFrequency,
        },
        { label: "Data rate", value: device.RX2Config.dataRate },
        { label: "ACK timeout", value: device.RX2Config.ACKTimeout },
      ],
    },
    {
      title: "Frame Configuration",
      icon: "data_usage",
      items: [
        { label: "FPort", value: device.frameConfig.fPort },
        {
          label: "Retransmission",
          value: device.frameConfig.retransmission,
        },
        { label: "FCnt up", value: displayValue(device.frameConfig.FCntUp) },
        {
          label: "FCnt down",
          value: displayValue(device.frameConfig.FCntDown),
        },
        {
          label: "Frame counter validation",
          value: displayBoolean(
            !device.frameConfig.disableFrameCounterValidation,
          ),
        },
      ],
    },
    {
      title: "Payload Configuration",
      icon: "data_object",
      layout: EntityDetailsSectionLayout.Full,
      items: [
        {
          label: "Uplink interval",
          value: device.payloadConfig.uplinkInterval,
        },
        {
          label: "Oversized payload",
          value: device.payloadConfig.oversizedPayloadBehavior,
          format: EntityDetailValueFormat.Enum,
        },
        {
          label: "Message type",
          value: device.payloadConfig.MType,
          format: EntityDetailValueFormat.Enum,
        },
        { label: "Payload", value: device.payloadConfig.payload },
        {
          label: "Base64 encoded",
          value: displayBoolean(device.payloadConfig.base64Encoded),
        },
      ],
    },
    {
      title: "Advanced Configuration",
      icon: "tune",
      items: [
        { label: "Antenna range", value: device.advancedConfig.antennaRange },
        {
          label: "Adaptive data rate",
          value: displayBoolean(device.advancedConfig.ADREnabled),
        },
      ],
    },
  ];

  if (device.OOTAConfig) {
    sections.splice(2, 0, {
      title: "OTAA Configuration",
      icon: "key",
      description: "Credentials used to join the LoRaWAN network",
      items: [
        { label: "JoinEUI", value: device.OOTAConfig.joinEUI, mono: true },
        { label: "AppKey", value: device.OOTAConfig.appKey, mono: true },
      ],
    });
  }

  if (device.ABPConfig) {
    sections.splice(2, 0, {
      title: "ABP Configuration",
      icon: "key",
      description: "Preconfigured session credentials",
      items: [
        { label: "DevAddr", value: device.ABPConfig.devAddr, mono: true },
        { label: "NwkSKey", value: device.ABPConfig.nwkSKey, mono: true },
        { label: "AppSKey", value: device.ABPConfig.appSKey, mono: true },
      ],
    });
  }

  return sections;
};

const DeviceInfoContent = () => {
  const deviceService = useDeviceService();
  const router = useRouter();
  const searchParams = useSearchParams();
  const [device, setDevice] = useState<Device | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const deviceId = searchParams.get("deviceId");

    if (!deviceId) {
      setError(new Error("Device identifier is missing"));
      setIsLoading(false);
      return;
    }

    let isCurrent = true;

    const loadDevice = async () => {
      try {
        const result = await deviceService.getDevice({ id: deviceId });

        if (isCurrent) {
          setDevice(result);
        }
      } catch (err) {
        if (isCurrent) {
          setError(
            err instanceof Error ? err : new Error("Failed to load device"),
          );
        }
      } finally {
        if (isCurrent) {
          setIsLoading(false);
        }
      }
    };

    loadDevice();

    return () => {
      isCurrent = false;
    };
  }, [deviceService, searchParams]);

  if (isLoading) {
    return (
      <div className="page device-page">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="page device-page">
      <PageHeader
        title={device?.name || "Device details"}
        subTitle="View device configuration and status"
      >
        <Button value="Back" onClick={() => router.push("/hardware/devices")} />
      </PageHeader>
      {error || !device ? (
        <p>Unable to load this device.</p>
      ) : (
        <EntityDetails sections={getDeviceSections(device)} />
      )}
    </div>
  );
};

const DevicePage = () => {
  return (
    <Suspense fallback={<Spinner />}>
      <DeviceInfoContent />
    </Suspense>
  );
};

export default DevicePage;
