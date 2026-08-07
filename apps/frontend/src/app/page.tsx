"use client";

import { useEffect, useState } from "react";
import { StatusResponse } from "@lwn-simulator/contracts";
import { Button, ButtonType, PageHeader } from "@lwn-simulator/ui-components";
import { SensorMap } from "@/components";

export default function HomePage() {
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [port, setPort] = useState<StatusResponse | null>(null);

  const getData = async (): Promise<void> => {
    setIsLoading(true);

    const response = await fetch(
      `${process.env.NODE_ENV === "development" ? "http://localhost:8080" : window.location.origin}/api/status`,
    );

    const data: StatusResponse = await response.json();

    console.log("Data", data);
    setPort(data);

    setIsLoading(false);
  };

  useEffect(() => {
    getData();
  }, []);

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
      {isLoading ? "Loading..." : `Running on port: ${port?.port}`}
    </div>
  );
}
