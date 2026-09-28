"use client";

import { Gateway, GatewayType } from "@lwn-simulator/contracts";
import { useGatewayService } from "@lwn-simulator/sdk";
import {
  Button,
  EntityDetailTone,
  EntityDetailSection,
  EntityDetails,
  EntityDetailValueFormat,
  PageHeader,
  Spinner,
} from "@lwn-simulator/ui-components";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";

const displayValue = (value: unknown) =>
  value === undefined || value === null || value === ""
    ? "Not configured"
    : String(value);

const getGatewaySections = (gateway: Gateway): Array<EntityDetailSection> => {
  const connectionItems =
    gateway.type === GatewayType.Virtual
      ? [
          {
            label: "Keep alive",
            value: `${displayValue(gateway.keepAlive)} seconds`,
          },
        ]
      : [
          { label: "Gateway IPv4", value: displayValue(gateway.gatewayIPv4) },
          { label: "Gateway port", value: displayValue(gateway.gatewayPort) },
        ];

  return [
    {
      title: "Overview",
      icon: "router",
      description: "Identity and current runtime state",
      items: [
        { label: "ID", value: gateway.id, mono: true },
        {
          label: "Status",
          value: gateway.active ? "Active" : "Inactive",
          tone: gateway.active
            ? EntityDetailTone.Positive
            : EntityDetailTone.Muted,
        },
        { label: "Name", value: gateway.name },
        {
          label: "Type",
          value: gateway.type,
          format: EntityDetailValueFormat.Enum,
        },
        { label: "MAC address", value: gateway.macAddress, mono: true },
        { label: "Gateway EUI", value: gateway.gatewayEUI, mono: true },
      ],
    },
    {
      title: "Connection",
      icon: "lan",
      description:
        gateway.type === GatewayType.Virtual
          ? "Virtual gateway heartbeat configuration"
          : "Network endpoint of the physical gateway",
      items: connectionItems,
    },
    {
      title: "Location",
      icon: "location_on",
      items: [
        { label: "Latitude", value: gateway.latitude },
        { label: "Longitude", value: gateway.longitude },
        { label: "Altitude", value: gateway.altitude },
      ],
    },
  ];
};

const GatewayInfoContent = () => {
  const gatewayService = useGatewayService();
  const router = useRouter();
  const searchParams = useSearchParams();
  const [gateway, setGateway] = useState<Gateway | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const gatewayId = searchParams.get("gatewayId");

    if (!gatewayId) {
      setError(new Error("Gateway identifier is missing"));
      setIsLoading(false);
      return;
    }

    let isCurrent = true;

    const loadGateway = async () => {
      try {
        const result = await gatewayService.getGateway({ id: gatewayId });

        if (isCurrent) {
          setGateway(result);
        }
      } catch (err) {
        if (isCurrent) {
          setError(
            err instanceof Error ? err : new Error("Failed to load gateway"),
          );
        }
      } finally {
        if (isCurrent) {
          setIsLoading(false);
        }
      }
    };

    loadGateway();

    return () => {
      isCurrent = false;
    };
  }, [gatewayService, searchParams]);

  if (isLoading) {
    return (
      <div className="page gateway-page">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="page gateway-page">
      <PageHeader
        title={gateway?.name || "Gateway details"}
        subTitle="View gateway configuration and status"
      >
        <Button
          value="Back"
          onClick={() => router.push("/hardware/gateways")}
        />
      </PageHeader>
      {error || !gateway ? (
        <p>Unable to load this gateway.</p>
      ) : (
        <EntityDetails sections={getGatewaySections(gateway)} />
      )}
    </div>
  );
};

const GatewayInfoPage = () => (
  <Suspense fallback={<Spinner />}>
    <GatewayInfoContent />
  </Suspense>
);

export default GatewayInfoPage;
