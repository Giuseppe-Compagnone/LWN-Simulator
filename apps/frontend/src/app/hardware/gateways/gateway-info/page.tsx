"use client";

import { Gateway } from "@lwn-simulator/contracts";
import { useGatewayService } from "@lwn-simulator/sdk";
import { Button, PageHeader, Spinner } from "@lwn-simulator/ui-components";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";

const GatewayInfoContent = () => {
  const gatewayService = useGatewayService();
  const router = useRouter();
  const searchParams = useSearchParams();
  const [gateway, setGateway] = useState<Gateway | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const gatewayId = searchParams.get("gatewayId");
    if (!gatewayId) {
      setIsLoading(false);
      return;
    }
    gatewayService.getGateway({ id: gatewayId }).then(setGateway).finally(() => setIsLoading(false));
  }, [gatewayService, searchParams]);

  if (isLoading) return <div className="page gateway-page"><Spinner /></div>;

  return (
    <div className="page gateway-page">
      <PageHeader title={gateway?.name || "Gateway details"} subTitle="View gateway configuration and status">
        <Button value="Back" onClick={() => router.push("/hardware/gateways")} />
      </PageHeader>
      {!gateway ? <p>Unable to load this gateway.</p> : (
        <section className="gateway-details">
          <p><strong>ID:</strong> {gateway.id}</p>
          <p><strong>Type:</strong> {gateway.type}</p>
          <p><strong>Status:</strong> {gateway.active ? "Active" : "Inactive"}</p>
          <p><strong>MAC address:</strong> {gateway.macAddress}</p>
          <p><strong>Gateway EUI:</strong> {gateway.gatewayEUI}</p>
          {gateway.type === "virtual" ? <p><strong>Keep alive:</strong> {gateway.keepAlive} seconds</p> : <>
            <p><strong>Gateway IPv4:</strong> {gateway.gatewayIPv4}</p>
            <p><strong>Gateway port:</strong> {gateway.gatewayPort}</p>
          </>}
          <p><strong>Latitude:</strong> {gateway.latitude}</p>
          <p><strong>Longitude:</strong> {gateway.longitude}</p>
          <p><strong>Altitude:</strong> {gateway.altitude}</p>
        </section>
      )}
    </div>
  );
};

const GatewayInfoPage = () => <Suspense fallback={<Spinner />}><GatewayInfoContent /></Suspense>;

export default GatewayInfoPage;
