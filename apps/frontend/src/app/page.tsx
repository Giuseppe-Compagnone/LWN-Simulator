"use client";

import { Button, ButtonType, Card, CardLayout, Spinner } from "@lwn-simulator/ui-components";
import { SimulationStatus } from "@lwn-simulator/contracts";
import { useDeviceService, useGatewayService, useSimulationService } from "@lwn-simulator/sdk";
import { SensorMap, useSensorMap } from "@/components";
import { useMemo, useState } from "react";

const formatCount = (value: number | null): string => value === null ? "—" : value.toString().padStart(2, "0");

const HomePage = () => {
  const [actionError, setActionError] = useState<string | null>(null);
  const deviceService = useDeviceService();
  const gatewayService = useGatewayService();
  const simulation = useSimulationService();
  const mapLogic = useSensorMap({});
  const devices = deviceService.devices;
  const gateways = gatewayService.gateways;
  const activeDevices = useMemo(() => devices?.filter((device) => device.active).length ?? null, [devices]);
  const activeGateways = useMemo(() => gateways?.filter((gateway) => gateway.active).length ?? null, [gateways]);
  const isLoading = devices === null || gateways === null;
  const isSimulationRunning = simulation.snapshot?.state.status === SimulationStatus.Running;
  const packetSuccessRate = simulation.snapshot?.metrics.packetSuccessRate;
  const averageSnr = simulation.snapshot?.metrics.averageSNR;
  const averageRssi = simulation.snapshot?.metrics.averageRSSI;
  const simulationStatus = simulation.snapshot?.state.status ?? SimulationStatus.Idle;
  const events = simulation.events.length > 0
    ? [...simulation.events].reverse().slice(0, 6).map((event) => ({ tone: "blue", label: event.type.toUpperCase(), message: event.message }))
    : [
      { tone: "neutral", label: "SYSTEM", message: isSimulationRunning ? "Simulation engine started" : "Simulation ready — waiting for start" },
      { tone: "blue", label: "DEVICES", message: devices === null ? "Loading device registry" : `${devices.length} devices loaded` },
      { tone: "green", label: "GATEWAYS", message: gateways === null ? "Loading gateway registry" : `${gateways.length} gateways available` },
    ];

  const startSimulation = async () => {
    try {
      setActionError(null);
      await simulation.start({ speed: 1 });
    } catch (error) {
      setActionError(error instanceof Error ? error.message : "Unable to start simulation");
    }
  };

  const stopSimulation = async () => {
    try {
      setActionError(null);
      await simulation.stop();
    } catch (error) {
      setActionError(error instanceof Error ? error.message : "Unable to stop simulation");
    }
  };

  return <main className="simulation-home page">
    <header className="simulation-home__header">
      <div><p className="simulation-home__eyebrow">LWN SIMULATOR / CONTROL CENTER</p><h1>Active Simulation Canvas</h1><p className="simulation-home__subtitle">Real-time LoRaWAN propagation modelling and network activity.</p></div>
      <div className="simulation-home__actions"><Button value="Stop" type={ButtonType.Outlined} disabled={!isSimulationRunning} onClick={stopSimulation} /><Button value="▶ Start Simulation" disabled={isSimulationRunning} onClick={startSimulation} /></div>
    </header>
    <div className="simulation-home__content">
      <section className="simulation-home__top-grid">
        <Card className="simulation-panel simulation-map-panel" layout={CardLayout.Padded}>
          <div className="simulation-panel__heading"><div><span className="simulation-panel__kicker">NETWORK TOPOLOGY</span><h2>Propagation map</h2></div><span className={`simulation-status ${isSimulationRunning ? "is-live" : "is-ready"}`}><span /> {isSimulationRunning ? "Live" : "Ready"}</span></div>
          <div className="simulation-map-panel__canvas"><SensorMap logic={mapLogic} devices={devices ?? []} gateways={gateways ?? []} /><div className="simulation-map-panel__overlay"><span className="material-symbols-outlined">map</span>Local network topology</div><div className="simulation-map-panel__legend"><span><i className="legend-dot legend-dot--device" /> Devices {formatCount(devices?.length ?? null)}</span><span><i className="legend-dot legend-dot--gateway" /> Gateways {formatCount(gateways?.length ?? null)}</span></div></div>
        </Card>
        <div className="simulation-home__metrics">
          <Card className="simulation-panel metric-card" layout={CardLayout.Padded}><div className="metric-card__topline"><span className="simulation-panel__kicker">PACKET SUCCESS RATE</span><span className="material-symbols-outlined metric-card__icon">monitoring</span></div><strong className="metric-card__value">{packetSuccessRate === undefined ? "—" : `${(packetSuccessRate * 100).toFixed(1)}%`}</strong><div className="metric-card__progress" aria-label="Packet success rate"><span style={{ width: `${(packetSuccessRate ?? 0) * 100}%` }} /></div><p className="metric-card__hint">{packetSuccessRate === undefined ? "Available when simulation traffic starts" : `${simulation.snapshot?.metrics.successfulUplinks ?? 0} successful uplinks`}</p></Card>
          <Card className="simulation-panel uptime-card" layout={CardLayout.Padded}><div className="metric-card__topline"><span className="simulation-panel__kicker">SIMULATION UPTIME</span><span className={`simulation-status simulation-status--soft ${isSimulationRunning ? "is-live" : ""}`}><span /> {simulationStatus}</span></div><div className="uptime-grid" aria-label="Simulation uptime">{Array.from({ length: 28 }, (_, index) => <span key={index} className={isSimulationRunning && index < 20 ? "is-active" : ""} />)}</div><div className="uptime-card__footer"><span>CONTINUOUS RUN</span><strong>{simulation.snapshot ? `${Math.floor(simulation.snapshot.state.elapsedMilliseconds / 3600000)}h` : "Not started"}</strong></div></Card>
        </div>
      </section>
      <section className="simulation-home__bottom-grid">
        <Card className="simulation-panel integrity-card" layout={CardLayout.Padded}><div className="simulation-panel__heading"><div><span className="simulation-panel__kicker">SIGNAL INTEGRITY</span><h2>Network health</h2></div><span className="material-symbols-outlined section-icon">wifi_tethering</span></div><div className="integrity-list"><div><span>SNR average</span><strong>{averageSnr === undefined ? "Awaiting telemetry" : `${averageSnr.toFixed(1)} dB`}</strong></div><div><span>RSSI floor</span><strong>{averageRssi === undefined ? "Awaiting telemetry" : `${averageRssi.toFixed(1)} dBm`}</strong></div><div><span>Error rate (BER)</span><strong className="is-muted">{simulation.snapshot ? `${(1 - (packetSuccessRate ?? 0)).toFixed(4)}` : "—"}</strong></div></div><div className="integrity-summary"><span className="material-symbols-outlined">hub</span><span><strong>{formatCount(activeDevices)} active nodes</strong> across <strong>{formatCount(activeGateways)} active gateways</strong></span></div></Card>
        <Card className="simulation-panel event-card" layout={CardLayout.Padded}><div className="simulation-panel__heading"><div><span className="simulation-panel__kicker">REAL-TIME EVENT STREAM</span><h2>System activity</h2></div><span className="event-filter">FILTER: ALL EVENTS <span className="material-symbols-outlined">tune</span></span></div><div className="event-list">{events.map((event, index) => <div className="event-row" key={`${event.label}-${index}`}><span className={`event-label event-label--${event.tone}`}>[{event.label}]</span><span>{event.message}</span></div>)}</div><div className="event-card__footer">{actionError ?? (isLoading ? <><Spinner /> Loading local registries</> : "Live stream will appear when the simulation starts")}</div></Card>
      </section>
    </div>
    <button className="simulation-home__quick-action" type="button" aria-label="Simulation actions"><span className="material-symbols-outlined">bolt</span></button>
  </main>;
};

export default HomePage;
