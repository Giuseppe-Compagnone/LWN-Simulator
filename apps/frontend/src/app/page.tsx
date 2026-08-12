"use client";

import { useEffect, useState } from "react";
import { Button, ButtonType, PageHeader } from "@lwn-simulator/ui-components";
import { SensorMap } from "@/components";
import { useAppInfoService } from "@lwn-simulator/sdk";

export default function HomePage() {
  // States
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [port, setPort] = useState<string | null>(null);

  // Hooks
  const appInfo = useAppInfoService();

  // Effects
  useEffect(() => {
    const getData = async (): Promise<void> => {
      setIsLoading(true);

      const data = await appInfo.status();
      console.log("Data", data);
      setPort(data.port);

      setIsLoading(false);
    };

    getData();
  }, [appInfo]);

  return (
    <div className="home-page page">
      <PageHeader
        title="Simulation Dashboard"
        subTitle="View results, scenarios and performance in real time"
      >
        <Button value={"Start Simulation"} />
        <Button value={"Stop"} type={ButtonType.Outlined} />
      </PageHeader>
      <SensorMap />
      {isLoading ? "Loading..." : `Running on port: ${port}`}
    </div>
  );
}
