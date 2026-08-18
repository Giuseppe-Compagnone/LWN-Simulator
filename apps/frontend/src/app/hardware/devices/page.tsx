"use client";

import { useDeviceService } from "@lwn-simulator/sdk";
import { Button, PageHeader, Table } from "@lwn-simulator/ui-components";
import { redirect } from "next/navigation";

const DevicesPage = () => {
  // Hooks
  const deviceService = useDeviceService();

  return (
    <div className="page devices-page">
      <PageHeader title={"Devices"} subTitle="Create and manage devices">
        <Button
          value={"Add Device"}
          onClick={() => {
            redirect("/hardware/devices/new");
          }}
        />
      </PageHeader>
      <Table
        rowLabels={[{ value: "name" }]}
        records={
          Array.isArray(deviceService.devices)
            ? deviceService.devices.map((device) => {
                return {
                  items: [
                    {
                      label: "name",
                      value: device.name,
                    },
                  ],
                };
              })
            : []
        }
        pageSize={8}
        isLoading={!Array.isArray(deviceService.devices)}
      />
    </div>
  );
};

export default DevicesPage;
