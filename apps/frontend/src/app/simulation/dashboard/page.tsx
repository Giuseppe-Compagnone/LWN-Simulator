"use client";

import { useEffect, useState } from "react";
import {
  Button,
  ButtonType,
  NotificationHandler,
  PageHeader,
} from "@lwn-simulator/ui-components";
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

      try {
        const data = await appInfoService.status();
        setPort(data.port);
      } catch {
        NotificationHandler.instance.error("Failed to load application status");
      } finally {
        setIsLoading(false);
      }
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
