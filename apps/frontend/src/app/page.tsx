"use client";

import { useEffect, useState } from "react";
import { Button, ButtonType, PageHeader } from "@lwn-simulator/ui-components";
import { SensorMap } from "@/components";
import { useAppInfoService, useDeviceService } from "@lwn-simulator/sdk";

export default function HomePage() {
  // States
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [port, setPort] = useState<string | null>(null);

  // Hooks
  const appInfoService = useAppInfoService();
  const deviceService = useDeviceService();

  // Effects
  useEffect(() => {
    const getData = async (): Promise<void> => {
      setIsLoading(true);

      const data = await appInfoService.status();
      console.log("Data", data);
      setPort(data.port);

      setIsLoading(false);
    };

    getData();
  }, [appInfoService]);

  return (
    <div className="home-page page">
      <PageHeader
        title="Simulation Dashboard"
        subTitle="View results, scenarios and performance in real time"
      >
        <Button
          value={"Start Simulation"}
          onClick={async () => {
            try {
              const device = await deviceService.createDevice({
                devEUI: "70B3D57ED0000001",
                latitude: 45.4642,
                longitude: 9.19,
              });

              console.log("DEVICE", device);
            } catch (err) {
              console.log(err);
            }
          }}
        />
        <Button value={"Stop"} type={ButtonType.Outlined} />
      </PageHeader>
      <SensorMap />
      {isLoading ? "Loading..." : `Running on port: ${port}`}
    </div>
  );
}
