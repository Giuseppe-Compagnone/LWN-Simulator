"use client";

import { useGatewayService } from "@lwn-simulator/sdk";
import { Button, NotificationHandler, PageHeader, Table } from "@lwn-simulator/ui-components";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

const GatewaysPage = () => {
  const gatewayService = useGatewayService();
  const router = useRouter();

  useEffect(() => {
    if (gatewayService.error) {
      NotificationHandler.instance.error("Failed to load gateways");
    }
  }, [gatewayService.error]);

  return (
    <div className="page gateways-page">
      <PageHeader title="Gateways" subTitle="Create and manage virtual and real gateways">
        <Button value="Add Gateway" onClick={() => router.push("/hardware/gateways/new")} />
      </PageHeader>
      <Table
        rowLabels={[{ value: "name" }, { value: "type" }, { value: "macAddress" }, { value: "gatewayEUI" }]}
        records={Array.isArray(gatewayService.gateways) ? gatewayService.gateways.map((gateway) => ({
          id: gateway.id,
          items: [
            { label: "name", value: gateway.name },
            { label: "type", value: gateway.type },
            { label: "macAddress", value: gateway.macAddress },
            { label: "gatewayEUI", value: gateway.gatewayEUI },
          ],
        })) : []}
        pageSize={6}
        onRowClick={(row) => {
          if (row.id) {
            router.push(`/hardware/gateways/gateway-info?gatewayId=${row.id}`);
          }
        }}
        isLoading={!Array.isArray(gatewayService.gateways)}
      />
    </div>
  );
};

export default GatewaysPage;
