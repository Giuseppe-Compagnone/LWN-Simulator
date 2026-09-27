"use client";

import { useDeviceService } from "@lwn-simulator/sdk";
import {
  Button,
  NotificationHandler,
  PageHeader,
  Table,
} from "@lwn-simulator/ui-components";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

const DevicesPage = () => {
  // Hooks
  const deviceService = useDeviceService();
  const router = useRouter();

  // Effects
  useEffect(() => {
    if (deviceService.error) {
      NotificationHandler.instance.error("Failed to load devices");
    }
  }, [deviceService.error]);

  return (
    <div className="page devices-page">
      <PageHeader title={"Devices"} subTitle="Create and manage devices">
        <Button
          value={"Add Device"}
          onClick={() => {
            router.push("/hardware/devices/new");
          }}
        />
      </PageHeader>
      <Table
        rowLabels={[{ value: "name" }, { value: "devEUI" }]}
        records={
          Array.isArray(deviceService.devices)
            ? deviceService.devices.map((device) => {
                return {
                  id: device.id,
                  items: [
                    {
                      label: "name",
                      value: device.name,
                    },
                    {
                      label: "devEUI",
                      value: device.devEUI,
                    },
                  ],
                };
              })
            : []
        }
        pageSize={6}
        onRowClick={(row) => {
          if (row.id) {
            router.push(`/hardware/devices/device-info?deviceId=${row.id}`);
          }
        }}
        isLoading={!Array.isArray(deviceService.devices)}
      />
    </div>
  );
};

export default DevicesPage;
