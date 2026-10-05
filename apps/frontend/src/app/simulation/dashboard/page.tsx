"use client";

import {
  Button,
  ButtonType,
  Card,
  CardLayout,
  Form,
  FormValue,
  NotificationHandler,
  Spinner,
  rangeField,
  textField,
} from "@lwn-simulator/ui-components";
import {
  GatewayType,
  SimulationEventType,
  SimulationStatus,
} from "@lwn-simulator/contracts";
import {
  useDeviceService,
  useGatewayService,
  useSimulationService,
} from "@lwn-simulator/sdk";
import {
  SensorMap,
  SimulationCommandPanel,
  SimulationSparkline,
  SensorMapLink,
  useSensorMap,
} from "@/components";
import { readGatewayBridgeConfig } from "@/utils/gatewayBridge";
import { useCallback, useEffect, useMemo, useState } from "react";

const formatEventType = (value: string): string =>
  value.replace(/-/g, " ").replace(/\b\w/g, (letter) => letter.toUpperCase());

const formatDuration = (milliseconds: number): string => {
  const totalSeconds = Math.floor(milliseconds / 1000);
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  return `${String(hours).padStart(2, "0")}h ${String(minutes).padStart(2, "0")}m ${String(seconds).padStart(2, "0")}s`;
};

const formatNumber = (value: number): string =>
  new Intl.NumberFormat("en-US", { maximumFractionDigits: 2 }).format(value);

const distanceMeters = (
  latitudeA: number,
  longitudeA: number,
  latitudeB: number,
  longitudeB: number,
): number => {
  const earthRadiusMeters = 6_371_000;
  const toRadians = (degrees: number): number => (degrees * Math.PI) / 180;
  const latitudeDelta = toRadians(latitudeB - latitudeA);
  const longitudeDelta = toRadians(longitudeB - longitudeA);
  const latitudeARadians = toRadians(latitudeA);
  const latitudeBRadians = toRadians(latitudeB);
  const haversine =
    Math.sin(latitudeDelta / 2) ** 2 +
    Math.cos(latitudeARadians) *
      Math.cos(latitudeBRadians) *
      Math.sin(longitudeDelta / 2) ** 2;

  return (
    2 *
    earthRadiusMeters *
    Math.asin(Math.sqrt(Math.min(1, Math.max(0, haversine))))
  );
};

const communicationEventTypes = new Set<SimulationEventType>([
  SimulationEventType.GatewayPacketReceived,
  SimulationEventType.JoinRequestReceived,
  SimulationEventType.DeviceUplinkTransmitted,
  SimulationEventType.DeviceDownlinkTransmitted,
  SimulationEventType.GatewayPacketIngress,
  SimulationEventType.GatewayPacketEgress,
  SimulationEventType.ClassBDownlinkTransmitted,
]);

const radioTransmissionEventTypes = new Set<SimulationEventType>([
  SimulationEventType.RadioTransmissionStarted,
]);

const linkActivityWindowMilliseconds = 15_000;

const SimulationDashboardPage = () => {
  const simulation = useSimulationService();
  const deviceService = useDeviceService();
  const gatewayService = useGatewayService();
  const devices = useMemo(
    () => deviceService.devices ?? [],
    [deviceService.devices],
  );
  const gateways = useMemo(
    () => gatewayService.gateways ?? [],
    [gatewayService.gateways],
  );
  const mapLogic = useSensorMap({});
  const [controlError, setControlError] = useState<string | null>(null);
  const [isLoadingSnapshot, setIsLoadingSnapshot] = useState(true);
  const getSnapshot = simulation.getSnapshot;
  const getEvents = simulation.getEvents;

  useEffect(() => {
    let mounted = true;

    void getSnapshot()
      .then(() => getEvents({ afterSequence: 0, limit: 1000 }))
      .catch(() => undefined)
      .finally(() => {
        if (mounted) setIsLoadingSnapshot(false);
      });

    return () => {
      mounted = false;
    };
  }, [getEvents, getSnapshot]);

  const snapshot = simulation.snapshot;
  const status = snapshot?.state.status ?? SimulationStatus.Idle;
  const isRunning = status === SimulationStatus.Running;
  const isPaused = status === SimulationStatus.Paused;
  const canStart =
    status === SimulationStatus.Idle ||
    status === SimulationStatus.Stopped ||
    status === SimulationStatus.Failed;
  const packetRate = snapshot?.metrics.packetSuccessRate ?? 0;
  const simulationEvents = simulation.events;
  const recentRates = useMemo(
    () =>
      simulationEvents
        .filter((event) => event.type === "metrics-updated")
        .map(() => packetRate * 100)
        .slice(-24),
    [packetRate, simulationEvents],
  );
  const visibleEvents = [...simulationEvents].reverse().slice(0, 14);
  const mapLinks = useMemo<Array<SensorMapLink>>(() => {
    const devicesByID = new Map(devices.map((device) => [device.id, device]));
    const gatewaysByID = new Map(
      gateways.map((gateway) => [gateway.id, gateway]),
    );
    const latestLinks = new Map<string, SensorMapLink & { sequence: number }>();
    const latestSimulationStartSequence = simulationEvents.reduce(
      (sequence, event) =>
        event.type === SimulationEventType.SimulationStarted
          ? Math.max(sequence, event.sequence)
          : sequence,
      0,
    );
    const elapsedMilliseconds = snapshot?.state.elapsedMilliseconds ?? 0;
    const minimumLinkTimestamp = Math.max(
      0,
      elapsedMilliseconds - linkActivityWindowMilliseconds,
    );

    const registerLink = (
      device: (typeof devices)[number],
      gateway: (typeof gateways)[number],
      sequence: number,
    ) => {
      if (!gateway.active) return;

      const distance = distanceMeters(
        device.locationConfig.latitude,
        device.locationConfig.longitude,
        gateway.latitude,
        gateway.longitude,
      );
      if (distance > device.advancedConfig.antennaRange) return;

      const id = `${device.id}:${gateway.id}`;
      const previous = latestLinks.get(id);
      if (previous && previous.sequence >= sequence) return;

      latestLinks.set(id, {
        id,
        sequence,
        from: {
          latitude: device.locationConfig.latitude,
          longitude: device.locationConfig.longitude,
        },
        to: {
          latitude: gateway.latitude,
          longitude: gateway.longitude,
        },
      });
    };

    for (const event of simulationEvents) {
      if (
        event.sequence < latestSimulationStartSequence ||
        event.timestampMilliseconds < minimumLinkTimestamp
      ) {
        continue;
      }

      if (!event.deviceID) continue;

      const device = devicesByID.get(event.deviceID);
      if (!device) continue;

      if (communicationEventTypes.has(event.type) && event.gatewayID) {
        const gateway = gatewaysByID.get(event.gatewayID);
        if (gateway) registerLink(device, gateway, event.sequence);
        continue;
      }

      if (!radioTransmissionEventTypes.has(event.type)) continue;

      for (const gateway of gateways) {
        if (gateway.type !== GatewayType.Virtual) continue;

        registerLink(device, gateway, event.sequence);
      }
    }

    return [...latestLinks.values()].map((link) => ({
      id: link.id,
      from: link.from,
      to: link.to,
    }));
  }, [devices, gateways, simulationEvents, snapshot?.state.elapsedMilliseconds]);

  const runControl = async (
    action: () => Promise<unknown>,
    successMessage: string,
  ) => {
    setControlError(null);
    try {
      await action();
      NotificationHandler.instance.success(successMessage);
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Simulation action failed";
      setControlError(message);
      NotificationHandler.instance.error(message);
    }
  };

  const start = (values: Record<string, FormValue>) => {
    const speed = Number(values.speed);
    const seed = typeof values.seed === "string" ? values.seed.trim() : "";
    const gatewayBridge = readGatewayBridgeConfig();

    return runControl(
      () =>
        simulation.start({
          speed,
        ...(seed ? { seed: Number(seed) } : {}),
        gatewayBridge,
      }),
      "Simulation started",
    );
  };

  const updateSpeed = useCallback(
    async (speed: number) => {
      try {
        await simulation.setSpeed({ speed });
        setControlError(null);
      } catch (error) {
        const message =
          error instanceof Error
            ? error.message
            : "Unable to update simulation speed";
        setControlError(message);
        NotificationHandler.instance.error(message);
      }
    },
    [simulation],
  );

  const controlFields = useMemo(
    () => [
      rangeField({
        name: "speed",
        label: "Speed",
        value: "1",
        error: null,
        min: 0.1,
        max: 10,
        step: 0.1,
        formatValue: (value: string) => `${Number(value).toFixed(1)}x`,
        onChange: (logic) => {
          const value = Number(logic.fieldsState.speed.value);

          if ((isRunning || isPaused) && Number.isFinite(value) && value > 0) {
            void updateSpeed(value);
          }
        },
      }),
      textField({
        name: "seed",
        label: "Seed (optional)",
        value: "",
        error: null,
        placeholder: "Optional",
        disabled: !canStart,
        format: (raw: string) => raw.replace(/[^0-9]/g, ""),
      }),
    ],
    [canStart, isPaused, isRunning, updateSpeed],
  );

  return (
    <main className="simulation-dashboard page">
      <header className="simulation-dashboard__header">
        <div>
          <p className="simulation-dashboard__eyebrow">
            LWN SIMULATOR / REALTIME ENGINE
          </p>
          <h1>Simulation control center</h1>
          <p className="simulation-dashboard__subtitle">
            Monitor propagation, protocol traffic and gateway connectivity in
            real time.
          </p>
        </div>
        <div className="simulation-dashboard__status">
          <span className={`simulation-status ${isRunning ? "is-live" : ""}`}>
            <span /> {formatEventType(status)}
          </span>
          <span className="simulation-dashboard__connection">
            WS {simulation.connectionState}
          </span>
        </div>
      </header>

      <Card
        className="simulation-dashboard__controls"
        layout={CardLayout.Padded}
      >
        <Form
          fields={controlFields}
          onSubmit={start}
          submitButton={{
            value: "Start simulation",
            className: `simulation-dashboard__start-button${canStart ? "" : " is-hidden"}`,
          }}
        />
        {(isRunning || isPaused) && (
          <div className="simulation-dashboard__control-actions">
            {isRunning && (
              <Button
                value="Pause"
                type={ButtonType.Outlined}
                onClick={() =>
                  runControl(simulation.pause, "Simulation paused")
                }
              />
            )}
            {isPaused && (
              <Button
                value="Resume"
                onClick={() =>
                  runControl(simulation.resume, "Simulation resumed")
                }
              />
            )}
            <Button
              value="Stop"
              type={ButtonType.Outlined}
              onClick={() => runControl(simulation.stop, "Simulation stopped")}
            />
          </div>
        )}
      </Card>

      {(controlError || simulation.error) && (
        <div className="simulation-dashboard__error" role="alert">
          {controlError ?? simulation.error?.message}
        </div>
      )}

      <section className="simulation-dashboard__metric-grid">
        <Card
          className="simulation-dashboard__metric"
          layout={CardLayout.Padded}
        >
          <span className="simulation-dashboard__kicker">
            PACKET SUCCESS RATE
          </span>
          <strong>
            {snapshot ? `${(packetRate * 100).toFixed(1)}%` : "—"}
          </strong>
          <SimulationSparkline
            values={recentRates}
            label="Packet success rate trend"
          />
          <small>
            {snapshot
              ? `${snapshot.metrics.successfulUplinks} successful uplinks`
              : "Start traffic to collect metrics"}
          </small>
        </Card>
        <Card
          className="simulation-dashboard__metric"
          layout={CardLayout.Padded}
        >
          <span className="simulation-dashboard__kicker">
            SIMULATION UPTIME
          </span>
          <strong>
            {snapshot
              ? formatDuration(snapshot.state.elapsedMilliseconds)
              : "—"}
          </strong>
          <small>
            {snapshot
              ? `${formatNumber(snapshot.state.speed)}x clock speed`
              : "Engine idle"}
          </small>
        </Card>
        <Card
          className="simulation-dashboard__metric"
          layout={CardLayout.Padded}
        >
          <span className="simulation-dashboard__kicker">RADIO HEALTH</span>
          <strong>
            {snapshot ? `${formatNumber(snapshot.metrics.averageSNR)} dB` : "—"}
          </strong>
          <small>
            {snapshot
              ? `${formatNumber(snapshot.metrics.averageRSSI)} dBm average RSSI`
              : "Awaiting telemetry"}
          </small>
        </Card>
        <Card
          className="simulation-dashboard__metric"
          layout={CardLayout.Padded}
        >
          <span className="simulation-dashboard__kicker">NETWORK RUNTIME</span>
          <strong>
            {snapshot
              ? `${snapshot.metrics.activeDevices} / ${snapshot.metrics.activeGateways}`
              : "—"}
          </strong>
          <small>active devices / gateways</small>
        </Card>
      </section>

      <SimulationCommandPanel
        devices={devices}
        enabled={isRunning || isPaused}
        onQueueUplink={simulation.queueUplink}
        onQueueDownlink={simulation.queueDownlink}
        onQueueMACCommand={simulation.queueMACCommand}
      />

      <section className="simulation-dashboard__main-grid">
        <Card className="simulation-dashboard__map" layout={CardLayout.Padded}>
          <div className="simulation-dashboard__section-heading">
            <div>
              <span className="simulation-dashboard__kicker">
                LIVE TOPOLOGY
              </span>
              <h2>Propagation map</h2>
            </div>
            <span>
              {devices.length} devices · {gateways.length} gateways
            </span>
          </div>
          <div className="simulation-dashboard__map-canvas">
            <SensorMap
              logic={mapLogic}
              devices={devices}
              gateways={gateways}
              links={mapLinks}
            />
          </div>
        </Card>
        <Card
          className="simulation-dashboard__events"
          layout={CardLayout.Padded}
        >
          <div className="simulation-dashboard__section-heading">
            <div>
              <span className="simulation-dashboard__kicker">
                REAL-TIME EVENT STREAM
              </span>
              <h2>Engine activity</h2>
            </div>
            <span>{simulation.events.length} events</span>
          </div>
          <div className="simulation-dashboard__event-list">
            {visibleEvents.length === 0 && (
              <p className="simulation-dashboard__empty">
                Events will appear when the engine starts.
              </p>
            )}
            {visibleEvents.map((event) => (
              <div className="simulation-dashboard__event" key={event.id}>
                <time>{formatDuration(event.timestampMilliseconds)}</time>
                <strong>{formatEventType(event.type)}</strong>
                <span>{event.message}</span>
              </div>
            ))}
          </div>
        </Card>
      </section>

      <section className="simulation-dashboard__runtime-grid">
        <Card layout={CardLayout.Padded}>
          <div className="simulation-dashboard__section-heading">
            <div>
              <span className="simulation-dashboard__kicker">
                DEVICE RUNTIME
              </span>
              <h2>Live sessions</h2>
            </div>
          </div>
          <div className="simulation-dashboard__runtime-list">
            {(snapshot?.devices ?? []).map((device) => (
              <div key={device.id}>
                <span>{device.name}</span>
                <strong>
                  {device.joined ? "Joined" : "Joining"} · FCnt{" "}
                  {device.frameCounterUp}
                </strong>
              </div>
            ))}
            {!snapshot?.devices.length && (
              <p className="simulation-dashboard__empty">No runtime devices.</p>
            )}
          </div>
        </Card>
        <Card layout={CardLayout.Padded}>
          <div className="simulation-dashboard__section-heading">
            <div>
              <span className="simulation-dashboard__kicker">
                GATEWAY RUNTIME
              </span>
              <h2>Connectivity</h2>
            </div>
          </div>
          <div className="simulation-dashboard__runtime-list">
            {(snapshot?.gateways ?? []).map((gateway) => (
              <div key={gateway.id}>
                <span>{gateway.name}</span>
                <strong>
                  {formatEventType(gateway.gatewayState)} ·{" "}
                  {gateway.ingressPackets} in / {gateway.egressPackets} out
                </strong>
              </div>
            ))}
            {!snapshot?.gateways.length && (
              <p className="simulation-dashboard__empty">
                No runtime gateways.
              </p>
            )}
          </div>
        </Card>
      </section>

      {isLoadingSnapshot && (
        <div className="simulation-dashboard__loading">
          <Spinner /> Loading simulation state…
        </div>
      )}
    </main>
  );
};

export default SimulationDashboardPage;
