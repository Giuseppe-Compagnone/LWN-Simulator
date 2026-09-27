"use client";

import { useDeviceService } from "@lwn-simulator/sdk";
import { Button, PageHeader, Spinner } from "@lwn-simulator/ui-components";
import { Device } from "@lwn-simulator/contracts";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";

const DeviceInfoContent = () => {
  // Hooks
  const deviceService = useDeviceService();
  const router = useRouter();
  const searchParams = useSearchParams();

  // States
  const [device, setDevice] = useState<Device | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  // Effects
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
        <section className="device-details">
          <p>
            <strong>ID:</strong> {device.id}
          </p>
          <p>
            <strong>DevEUI:</strong> {device.devEUI}
          </p>
          <p>
            <strong>Activation:</strong> {device.activation}
          </p>
          <p>
            <strong>Class:</strong> {device.class}
          </p>
          <p>
            <strong>Status:</strong> {device.active ? "Active" : "Inactive"}
          </p>
        </section>
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
