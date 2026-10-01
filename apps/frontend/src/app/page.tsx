"use client";

import { Button, ButtonType, Card, CardLayout, Spinner } from "@lwn-simulator/ui-components";
import { useDeviceService, useGatewayService } from "@lwn-simulator/sdk";
import { SensorMap, useSensorMap } from "@/components";
import { useMemo, useState } from "react";

const formatCount = (value: number | null): string => value === null ? "—" : value.toString().padStart(2, "0");

const HomePage = () => {
  const [isSimulationRunning, setIsSimulationRunning] = useState(false);
  const deviceService = useDeviceService();
  const gatewayService = useGatewayService();
  const mapLogic = useSensorMap({});
  const devices = deviceService.devices;
  const gateways = gatewayService.gateways;
  const activeDevices = useMemo(() => devices?.filter((device) => device.active).length ?? null, [devices]);
  const activeGateways = useMemo(() => gateways?.filter((gateway) => gateway.active).length ?? null, [gateways]);
  const isLoading = devices === null || gateways === null;
  const events = [
    { tone: "neutral", label: "SYSTEM", message: isSimulationRunning ? "Simulation engine started" : "Simulation ready — waiting for start" },
    { tone: "blue", label: "DEVICES", message: devices === null ? "Loading device registry" : `${devices.length} devices loaded` },
    { tone: "green", label: "GATEWAYS", message: gateways === null ? "Loading gateway registry" : `${gateways.length} gateways available` },
  ];

  return <main className="simulation-home page">
    <header className="simulation-home__header">
      <div><p className="simulation-home__eyebrow">LWN SIMULATOR / CONTROL CENTER</p><h1>Active Simulation Canvas</h1><p className="simulation-home__subtitle">Real-time LoRaWAN propagation modelling and network activity.</p></div>
      <div className="simulation-home__actions"><Button value="Stop" type={ButtonType.Outlined} disabled={!isSimulationRunning} onClick={() => setIsSimulationRunning(false)} /><Button value="▶ Start Simulation" onClick={() => setIsSimulationRunning(true)} /></div>
    </header>
    <div className="simulation-home__content">
      <section className="simulation-home__top-grid">
        <Card className="simulation-panel simulation-map-panel" layout={CardLayout.Padded}>
          <div className="simulation-panel__heading"><div><span className="simulation-panel__kicker">NETWORK TOPOLOGY</span><h2>Propagation map</h2></div><span className={`simulation-status ${isSimulationRunning ? "is-live" : "is-ready"}`}><span /> {isSimulationRunning ? "Live" : "Ready"}</span></div>
          <div className="simulation-map-panel__canvas"><SensorMap logic={mapLogic} devices={devices ?? []} gateways={gateways ?? []} /><div className="simulation-map-panel__overlay"><span className="material-symbols-outlined">map</span>Local network topology</div><div className="simulation-map-panel__legend"><span><i className="legend-dot legend-dot--device" /> Devices {formatCount(devices?.length ?? null)}</span><span><i className="legend-dot legend-dot--gateway" /> Gateways {formatCount(gateways?.length ?? null)}</span></div></div>
        </Card>
        <div className="simulation-home__metrics">
          <Card className="simulation-panel metric-card" layout={CardLayout.Padded}><div className="metric-card__topline"><span className="simulation-panel__kicker">PACKET SUCCESS RATE</span><span className="material-symbols-outlined metric-card__icon">monitoring</span></div><strong className="metric-card__value">—<small>%</small></strong><div className="metric-card__progress"><span /></div><p className="metric-card__hint">Available when simulation traffic starts</p></Card>
          <Card className="simulation-panel uptime-card" layout={CardLayout.Padded}><div className="metric-card__topline"><span className="simulation-panel__kicker">SIMULATION UPTIME</span><span className="simulation-status simulation-status--soft"><span /> Idle</span></div><div className="uptime-grid" aria-label="Simulation uptime unavailable">{Array.from({ length: 28 }, (_, index) => <span key={index} className={isSimulationRunning ? "is-active" : ""} />)}</div><div className="uptime-card__footer"><span>CONTINUOUS RUN</span><strong>{isSimulationRunning ? "00d 00h 00m" : "Not started"}</strong></div></Card>
        </div>
      </section>
      <section className="simulation-home__bottom-grid">
        <Card className="simulation-panel integrity-card" layout={CardLayout.Padded}><div className="simulation-panel__heading"><div><span className="simulation-panel__kicker">SIGNAL INTEGRITY</span><h2>Network health</h2></div><span className="material-symbols-outlined section-icon">wifi_tethering</span></div><div className="integrity-list"><div><span>SNR average</span><strong>Awaiting telemetry</strong></div><div><span>RSSI floor</span><strong>Awaiting telemetry</strong></div><div><span>Error rate (BER)</span><strong className="is-muted">—</strong></div></div><div className="integrity-summary"><span className="material-symbols-outlined">hub</span><span><strong>{formatCount(activeDevices)} active nodes</strong> across <strong>{formatCount(activeGateways)} active gateways</strong></span></div></Card>
        <Card className="simulation-panel event-card" layout={CardLayout.Padded}><div className="simulation-panel__heading"><div><span className="simulation-panel__kicker">REAL-TIME EVENT STREAM</span><h2>System activity</h2></div><span className="event-filter">FILTER: ALL EVENTS <span className="material-symbols-outlined">tune</span></span></div><div className="event-list">{events.map((event) => <div className="event-row" key={event.label}><span className={`event-label event-label--${event.tone}`}>[{event.label}]</span><span>{event.message}</span></div>)}</div><div className="event-card__footer">{isLoading ? <><Spinner /> Loading local registries</> : "Live stream will appear when the simulation starts"}</div></Card>
      </section>
    </div>
    <button className="simulation-home__quick-action" type="button" aria-label="Simulation actions"><span className="material-symbols-outlined">bolt</span></button>
  </main>;
};

export default HomePage;
