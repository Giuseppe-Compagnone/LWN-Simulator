"use client";

import { useEffect, useState } from "react";
import { Button, ButtonType, PageHeader } from "@lwn-simulator/ui-components";
import { SensorMap, useSensorMap } from "@/components";
import { useAppInfoService } from "@lwn-simulator/sdk";

const DashboardPage = () => {
  // States
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [port, setPort] = useState<string | null>(null);

  // Hooks
  const appInfoService = useAppInfoService();
  const mapLogic = useSensorMap({});

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
        <Button value={"Start Simulation"} />
        <Button value={"Stop"} type={ButtonType.Outlined} />
      </PageHeader>
      <SensorMap logic={mapLogic} />
      {isLoading ? "Loading..." : `Running on port: ${port}`}
    </div>
  );
};

export default DashboardPage;
