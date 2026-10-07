"use client";

import { RuntimeDevice, RuntimeGateway } from "@lwn-simulator/contracts";
import { memo, ReactNode, UIEvent, useMemo, useState } from "react";

const runtimeRowHeight = 52;
const runtimeViewportHeight = 320;
const runtimeOverscan = 4;

interface VirtualRuntimeListProps<T extends { id: string }> {
  items: Array<T>;
  emptyMessage: string;
  renderItem: (item: T) => ReactNode;
}

const VirtualRuntimeList = <T extends { id: string }>(
  props: VirtualRuntimeListProps<T>,
) => {
  const [scrollTop, setScrollTop] = useState(0);
  const range = useMemo(() => {
    const firstVisible = Math.floor(scrollTop / runtimeRowHeight);
    const visibleCount = Math.ceil(
      runtimeViewportHeight / runtimeRowHeight,
    );

    return {
      start: Math.max(0, firstVisible - runtimeOverscan),
      end: Math.min(
        props.items.length,
        firstVisible + visibleCount + runtimeOverscan,
      ),
    };
  }, [props.items.length, scrollTop]);

  if (props.items.length === 0) {
    return <p className="simulation-dashboard__empty">{props.emptyMessage}</p>;
  }

  return (
    <div
      className="simulation-dashboard__runtime-list"
      onScroll={(event: UIEvent<HTMLDivElement>) =>
        setScrollTop(event.currentTarget.scrollTop)
      }
    >
      <div
        className="simulation-dashboard__runtime-list-space"
        style={{ height: props.items.length * runtimeRowHeight }}
      >
        {props.items.slice(range.start, range.end).map((item, offset) => {
          const index = range.start + offset;

          return (
            <div
              className="simulation-dashboard__runtime-row"
              key={item.id}
              style={{ transform: `translateY(${index * runtimeRowHeight}px)` }}
            >
              {props.renderItem(item)}
            </div>
          );
        })}
      </div>
    </div>
  );
};

export const DeviceRuntimeList = memo(
  ({ devices }: { devices: Array<RuntimeDevice> }) => (
    <VirtualRuntimeList
      items={devices}
      emptyMessage="No runtime devices."
      renderItem={(device) => (
        <>
          <span>{device.name}</span>
          <strong>
            {device.joined ? "Joined" : "Joining"} · FCnt {device.frameCounterUp}
          </strong>
        </>
      )}
    />
  ),
);

DeviceRuntimeList.displayName = "DeviceRuntimeList";

export const GatewayRuntimeList = memo(
  ({
    gateways,
    formatState,
  }: {
    gateways: Array<RuntimeGateway>;
    formatState: (value: string) => string;
  }) => (
    <VirtualRuntimeList
      items={gateways}
      emptyMessage="No runtime gateways."
      renderItem={(gateway) => (
        <>
          <span>{gateway.name}</span>
          <strong>
            {formatState(gateway.gatewayState)} · {gateway.ingressPackets} in /{" "}
            {gateway.egressPackets} out
          </strong>
        </>
      )}
    />
  ),
);

GatewayRuntimeList.displayName = "GatewayRuntimeList";
