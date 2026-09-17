"use client";

import { useDeviceService } from "@lwn-simulator/sdk";
import {
  Button,
  NotificationHandler,
  PageHeader,
  Table,
} from "@lwn-simulator/ui-components";
import { redirect, useRouter } from "next/navigation";

const DevicesPage = () => {
  // Hooks
  const deviceService = useDeviceService();
  const router = useRouter();

  return (
    <div className="page devices-page">
      <PageHeader title={"Devices"} subTitle="Create and manage devices">
        <Button
          value={"Add Device"}
          onClick={() => {
            console.log("A");
            NotificationHandler.instance.success("AAA");
            // redirect("/hardware/devices/new");
          }}
        />
      </PageHeader>
      <Table
        rowLabels={[{ value: "name" }, { value: "devEUI" }]}
        records={
          Array.isArray(deviceService.devices)
            ? deviceService.devices.map((device) => {
                return {
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
          const id = row.items.find((i) => i.label == "devEUI");

          if (id)
            router.push(`/hardware/devices/device-info?deviceId=${id.value}`);
        }}
        isLoading={!Array.isArray(deviceService.devices)}
      />
    </div>
  );
};

export default DevicesPage;
