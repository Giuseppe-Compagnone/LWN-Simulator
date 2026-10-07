"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Suspense, useEffect, useMemo, useRef, useState } from "react";
import {
  SimulationEvent,
  SimulationLogObject,
  SimulationRun,
} from "@lwn-simulator/contracts";
import { useSimulationService } from "@lwn-simulator/sdk";
import { Button, Spinner } from "@lwn-simulator/ui-components";
import "../logs.scss";

const formatDate = (value: number): string =>
  new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));

const formatEventType = (value: string): string =>
  value.replace(/-/g, " ").replace(/\b\w/g, (letter) => letter.toUpperCase());

const LogDetailsContent = () => {
  const simulation = useSimulationService();
  const getLog = simulation.getLog;
  const getLogEvents = simulation.getLogEvents;
  const runID = useSearchParams().get("runId");
  const [run, setRun] = useState<SimulationRun | null>(null);
  const [events, setEvents] = useState<SimulationEvent[]>([]);
  const [selectedObject, setSelectedObject] =
    useState<SimulationLogObject | null>(null);
  const eventListRef = useRef<HTMLDivElement>(null);
  const eventListWasNearEnd = useRef(true);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!runID) {
      setError("No simulation run was selected.");
      setLoading(false);
      return;
    }

    let mounted = true;
    let firstLoad = true;

    const refreshLog = async () => {
      try {
        const [loadedRun, loadedEvents] = await Promise.all([
          getLog(runID),
          getLogEvents(runID),
        ]);
        if (!mounted) return;
        setRun(loadedRun);
        setEvents(loadedEvents.events);
        setSelectedObject((currentObject) =>
          currentObject ?? loadedRun.objects[0] ?? null,
        );
        setError(null);
      } catch {
        if (mounted && firstLoad) setError("Unable to load this simulation log.");
      } finally {
        if (mounted && firstLoad) {
          firstLoad = false;
          setLoading(false);
        }
      }
    };

    void refreshLog();
    const refreshInterval = window.setInterval(() => void refreshLog(), 1000);

    return () => {
      mounted = false;
      window.clearInterval(refreshInterval);
    };
  }, [getLog, getLogEvents, runID]);

  const objectEvents = useMemo(
    () =>
      selectedObject
        ? events.filter(
            (event) =>
              event.deviceID === selectedObject.id ||
              event.gatewayID === selectedObject.id,
          )
        : [],
    [events, selectedObject],
  );

  useEffect(() => {
    const eventList = eventListRef.current;
    if (!eventList || !eventListWasNearEnd.current) return;

    eventList.scrollTo({
      top: eventList.scrollHeight,
      behavior: "smooth",
    });
  }, [events]);

  if (loading) {
    return (
      <div className="logs-page__loading">
        <Spinner />
      </div>
    );
  }

  if (error || !run) {
    return (
      <section className="logs-page">
        <p className="logs-page__empty">{error ?? "Simulation log not found."}</p>
      </section>
    );
  }

  return (
    <section className="logs-page logs-page--detail">
      <header className="logs-page__header">
        <div>
          <span className="eyebrow">Simulation run</span>
          <h1>{formatDate(run.summary.startedAtMilliseconds)}</h1>
          <p>
            {run.summary.eventCount} recorded events · {run.summary.deviceCount} devices · {run.summary.gatewayCount} gateways
          </p>
        </div>
        <div className="logs-page__header-actions">
          <span className={`logs-page__status logs-page__status--${run.summary.status}`}>
            {formatEventType(run.summary.status)}
          </span>
          <Link href="/logs">
            <Button value="Back" />
          </Link>
        </div>
      </header>
      <div className="logs-page__detail-grid">
        <div className="logs-page__objects-column">
          <section className="logs-page__panel logs-page__panel--objects">
            <div className="logs-page__panel-head">
              <div>
                <span className="eyebrow">Object history</span>
                <h2>Devices and gateways</h2>
              </div>
            </div>
            <div className="logs-page__objects">
              {run.objects.map((object) => (
                <button
                  className={selectedObject?.id === object.id ? "logs-page__object logs-page__object--active" : "logs-page__object"}
                  key={object.id}
                  onClick={() => setSelectedObject(object)}
                  type="button"
                >
                  <span>{object.name}</span>
                  <small>
                    {formatEventType(object.kind)} · {object.identifier}
                  </small>
                </button>
              ))}
            </div>
          </section>
          <section className="logs-page__panel logs-page__panel--object-events">
            <div className="logs-page__panel-head">
              <span className="eyebrow">Selected object</span>
              <h2>{selectedObject?.name ?? "Object events"}</h2>
            </div>
            <div className="logs-page__object-events">
              {!selectedObject || objectEvents.length === 0 ? (
                <p>No events for this object.</p>
              ) : (
                objectEvents.map((event) => (
                  <div className="logs-page__event" key={event.id}>
                    <time>#{event.sequence}</time>
                    <span>{formatEventType(event.type)}</span>
                    <small>{event.message}</small>
                  </div>
                ))
              )}
            </div>
          </section>
        </div>
        <section className="logs-page__panel logs-page__panel--events">
          <span className="eyebrow">Event stream</span>
          <h2>All recorded events</h2>
          <div
            className="logs-page__event-list"
            onScroll={(event) => {
              const eventList = event.currentTarget;
              eventListWasNearEnd.current =
                eventList.scrollHeight -
                  eventList.scrollTop -
                  eventList.clientHeight <=
                96;
            }}
            ref={eventListRef}
          >
            {events.length === 0 ? (
              <p>No events were recorded.</p>
            ) : (
              events.map((event) => (
                <div className="logs-page__event" key={event.id}>
                  <time>#{event.sequence}</time>
                  <span>{formatEventType(event.type)}</span>
                  <small>{event.message}</small>
                </div>
              ))
            )}
          </div>
        </section>
      </div>
    </section>
  );
};

const LogDetailsPage = () => (
  <Suspense
    fallback={
      <div className="logs-page__loading">
        <Spinner />
      </div>
    }
  >
    <LogDetailsContent />
  </Suspense>
);

export default LogDetailsPage;
