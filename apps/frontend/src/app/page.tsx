/* eslint-disable react-hooks/set-state-in-effect */
"use client";

import { useEffect, useState } from "react";
import { Response } from "@lwn-simulator/contracts";

export default function Home() {
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [time, setTime] = useState<Response | null>(null);

  const getData = async (): Promise<void> => {
    setIsLoading(true);

    const response = await fetch(
      `http://${window.location.hostname}:8080/api/status`,
    );

    const data: Response = await response.json();

    console.log("Data", data);
    setTime(data);

    setIsLoading(false);
  };

  useEffect(() => {
    getData();
  }, []);

  return <div>{isLoading ? "Loading..." : time?.time}</div>;
}
