"use client";

import { DeviceForm } from "@/components/Hardware";
import {
  CreateDeviceRequest,
  Device,
  DeviceActivation,
  DeviceClass,
  DeviceMType,
  DeviceRegion,
  OversizedPayloadBehavior,
} from "@lwn-simulator/contracts";
import { useDeviceService } from "@lwn-simulator/sdk";
import {
  FormValue,
  NotificationHandler,
  PageHeader,
  Spinner,
} from "@lwn-simulator/ui-components";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";

const NewDevicePage = () => {
  // States
  const [editingDevice, setEditingDevice] = useState<Device | null>(null);
  const [isLoadingEdit, setIsLoadingEdit] = useState(false);

  // Hooks
  const deviceService = useDeviceService();
  const router = useRouter();
  const searchParams = useSearchParams();
  const deviceId = searchParams.get("deviceId");

  // Effects
  useEffect(() => {
    if (!deviceId) {
      setEditingDevice(null);
      setIsLoadingEdit(false);
      return;
    }

    let isCurrent = true;
    setIsLoadingEdit(true);
    deviceService
      .getDevice({ id: deviceId })
      .then((device) => {
        if (isCurrent) setEditingDevice(device);
      })
      .catch(() => {
        if (isCurrent)
          NotificationHandler.instance.error("Failed to load device");
      })
      .finally(() => {
        if (isCurrent) setIsLoadingEdit(false);
      });

    return () => {
      isCurrent = false;
    };
  }, [deviceId, deviceService]);

  // Callbacks
  const handleSubmit = async (values: Record<string, FormValue>) => {
    const req: CreateDeviceRequest = {
      devEUI: values.devEUI as string,
      name: values.name as string,
      activation: values.activation as DeviceActivation,
      class: values.class as DeviceClass,
      locationConfig: {
        latitude: Number(values.latitude),
        longitude: Number(values.longitude),
        altitude: Number(values.altitude),
        region: values.region as DeviceRegion,
      },
      RX1Config: {
        delay: Number(values.rx1Delay),
        duration: Number(values.rx1Duration),
        dataRateOffset: Number(values.rx1DRO),
      },
      RX2Config: {
        delay: Number(values.rx2Delay),
        duration: Number(values.rx2Duration),
        channelFrequency: Number(values.rx2CF),
        dataRate: values.rx2DR as number,
        ACKTimeout: Number(values.rx2ACKT),
      },
      advancedConfig: {
        antennaRange: Number(values.antennaRange),
        ADREnabled: values.ADREnabled as boolean,
      },
      frameConfig: {
        fPort: Number(values.fPort),
        retransmission: Number(values.retransmission),
        disableFrameCounterValidation:
          values.disableFrameCounterValidation as boolean,
        FCntUp: values.fCntUp ? Number(values.fCntUp) : undefined,
        FCntDown: values.fCntDown ? Number(values.fCntDown) : undefined,
      },
      payloadConfig: {
        uplinkInterval: Number(values.uplinkInterval),
        oversizedPayloadBehavior:
          values.oversizedPayloadBehavior as OversizedPayloadBehavior,
        MType: values.MType as DeviceMType,
        payload: values.payload as string,
        base64Encoded: values.base64Encoded as boolean,
      },
      ...((values.activation as DeviceActivation) == DeviceActivation.Oota
        ? {
            OOTAConfig: {
              joinEUI: values.joinEUI as string,
              appKey: values.appKey as string,
            },
          }
        : {}),
      ...((values.activation as DeviceActivation) == DeviceActivation.Abp
        ? {
            ABPConfig: {
              devAddr: values.devAddr as string,
              nwkSKey: values.nwkSKey as string,
              appSKey: values.appSKey as string,
            },
          }
        : {}),
    };

    try {
      if (editingDevice) {
        await deviceService.updateDevice({
          id: editingDevice.id,
          device: {
            ...editingDevice,
            ...req,
            devEUI: editingDevice.devEUI,
            class: editingDevice.class,
            activation: editingDevice.activation,
            active: values.active as boolean,
          },
        });
        NotificationHandler.instance.success("Device updated");
        router.back();
      } else {
        await deviceService.createDevice(req);
        NotificationHandler.instance.success("Device created");
        router.push("/hardware/devices");
      }
    } catch {
      NotificationHandler.instance.error(
        editingDevice ? "Failed to update device" : "Failed to create device",
      );
    }
  };

  if (isLoadingEdit)
    return (
      <div className="page entity-details-loading">
        <Spinner />
      </div>
    );

  return (
    <div className="page new-device-page">
      <PageHeader title={editingDevice ? "Edit Device" : "Create new Device"} />
      <DeviceForm device={editingDevice} onSubmit={handleSubmit} />
    </div>
  );
};

const NewDevicePageWithSuspense = () => (
  <Suspense
    fallback={
      <div className="page entity-details-loading">
        <Spinner />
      </div>
    }
  >
    <NewDevicePage />
  </Suspense>
);

export default NewDevicePageWithSuspense;
