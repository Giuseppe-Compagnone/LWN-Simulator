"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import {
  CSSProperties,
  Suspense,
  UIEvent,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  SimulationEvent,
  SimulationLogObject,
  SimulationRun,
} from "@lwn-simulator/contracts";
import { useSimulationService } from "@lwn-simulator/sdk";
import { Button, Spinner } from "@lwn-simulator/ui-components";
import "../logs.scss";

const eventPageSize = 1000;
const maxLiveEvents = 5000;
const eventRowHeight = 76;
const objectRowHeight = 76;
const virtualOverscan = 5;

const formatDate = (value: number): string =>
  new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));

const formatEventType = (value: string): string =>
  value.replace(/-/g, " ").replace(/\b\w/g, (letter) => letter.toUpperCase());

const formatSimulationTime = (milliseconds: number): string => {
  const totalSeconds = Math.max(0, milliseconds) / 1000;
  return `t+${totalSeconds.toFixed(totalSeconds >= 100 ? 0 : 1)}s`;
};

const mergeEvents = (
  current: SimulationEvent[],
  incoming: SimulationEvent[],
  maxEvents?: number,
): SimulationEvent[] => {
  const bySequence = new Map<number, SimulationEvent>();
  [...current, ...incoming].forEach((event) => bySequence.set(event.sequence, event));
  const merged = [...bySequence.values()].sort((left, right) => left.sequence - right.sequence);
  return maxEvents && merged.length > maxEvents ? merged.slice(-maxEvents) : merged;
};

interface VirtualEventListProps {
  events: SimulationEvent[];
  emptyMessage: string;
  autoScroll: boolean;
  onReachStart?: () => void;
  onNearEnd?: (nearEnd: boolean) => void;
}

const VirtualEventList = ({
  events,
  emptyMessage,
  autoScroll,
  onReachStart,
  onNearEnd,
}: VirtualEventListProps) => {
  const listRef = useRef<HTMLDivElement>(null);
  const [scrollTop, setScrollTop] = useState(0);
  const [viewportHeight, setViewportHeight] = useState(320);
  const firstVisible = Math.floor(scrollTop / eventRowHeight);
  const visibleCount = Math.ceil(viewportHeight / eventRowHeight);
  const start = Math.max(0, firstVisible - virtualOverscan);
  const end = Math.min(
    events.length,
    firstVisible + visibleCount + virtualOverscan,
  );

  useEffect(() => {
    const list = listRef.current;
    if (!list) return;
    const observer = new ResizeObserver(() => setViewportHeight(list.clientHeight || 320));
    observer.observe(list);
    setViewportHeight(list.clientHeight || 320);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    const list = listRef.current;
    if (list && autoScroll) list.scrollTop = list.scrollHeight;
  }, [autoScroll, events.length]);

  if (events.length === 0) return <p>{emptyMessage}</p>;

  return (
    <div
      className="logs-page__virtual-list"
      onScroll={(event: UIEvent<HTMLDivElement>) => {
        const list = event.currentTarget;
        const nearEnd = list.scrollHeight - list.scrollTop - list.clientHeight <= 96;
        setScrollTop(list.scrollTop);
        onNearEnd?.(nearEnd);
        if (list.scrollTop <= 96) onReachStart?.();
      }}
      ref={listRef}
    >
      <div style={{ height: events.length * eventRowHeight, position: "relative" }}>
        {events.slice(start, end).map((event, offset) => {
          const index = start + offset;
          const style: CSSProperties = {
            height: eventRowHeight,
            transform: `translateY(${index * eventRowHeight}px)`,
          };
          return (
            <div className="logs-page__event" key={event.id} style={style}>
              <time>
                #{event.sequence}
                <small>{formatSimulationTime(event.timestampMilliseconds)}</small>
              </time>
              <span>{formatEventType(event.type)}</span>
              <small>
                {event.message}
                {event.gatewayID ? ` · Gateway ${event.gatewayID}` : ""}
                {event.deviceID ? ` · Device ${event.deviceID}` : ""}
              </small>
            </div>
          );
        })}
      </div>
    </div>
  );
};

interface VirtualObjectListProps {
  objects: SimulationLogObject[];
  selectedObject: SimulationLogObject | null;
  onSelect: (object: SimulationLogObject) => void;
}

const VirtualObjectList = ({
  objects,
  selectedObject,
  onSelect,
}: VirtualObjectListProps) => {
  const [scrollTop, setScrollTop] = useState(0);
  const firstVisible = Math.floor(scrollTop / objectRowHeight);
  const start = Math.max(0, firstVisible - virtualOverscan);
  const end = Math.min(
    objects.length,
    firstVisible + Math.ceil(320 / objectRowHeight) + virtualOverscan,
  );

  useEffect(() => setScrollTop(0), [objects]);

  if (objects.length === 0) return <p>No matching devices or gateways.</p>;

  return (
    <div
      className="logs-page__objects"
      onScroll={(event) => setScrollTop(event.currentTarget.scrollTop)}
    >
      <div style={{ height: objects.length * objectRowHeight, position: "relative" }}>
        {objects.slice(start, end).map((object, offset) => {
          const index = start + offset;
          return (
            <button
              className={selectedObject?.id === object.id ? "logs-page__object logs-page__object--active" : "logs-page__object"}
              key={object.id}
              onClick={() => onSelect(object)}
              style={{ transform: `translateY(${index * objectRowHeight}px)` }}
              type="button"
            >
              <span>{object.name}</span>
              <small>{formatEventType(object.kind)} · {object.identifier}</small>
            </button>
          );
        })}
      </div>
    </div>
  );
};

const LogDetailsContent = () => {
  const simulation = useSimulationService();
  const getLog = simulation.getLog;
  const getLogEvents = simulation.getLogEvents;
  const runID = useSearchParams().get("runId");
  const [run, setRun] = useState<SimulationRun | null>(null);
  const runRef = useRef<SimulationRun | null>(null);
  const [events, setEvents] = useState<SimulationEvent[]>([]);
  const eventsRef = useRef<SimulationEvent[]>([]);
  const latestSequenceRef = useRef(0);
  const [selectedObject, setSelectedObject] = useState<SimulationLogObject | null>(null);
  const [objectEvents, setObjectEvents] = useState<SimulationEvent[]>([]);
  const objectEventsRef = useRef<SimulationEvent[]>([]);
  const objectLatestSequenceRef = useRef(0);
  const [objectLoading, setObjectLoading] = useState(false);
  const [loadingOlderObjectEvents, setLoadingOlderObjectEvents] = useState(false);
  const [objectFilter, setObjectFilter] = useState("");
  const [eventListNearEnd, setEventListNearEnd] = useState(true);
  const [loadingOlderEvents, setLoadingOlderEvents] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const updateEvents = useCallback((incoming: SimulationEvent[], maxEvents?: number) => {
    setEvents((current) => {
      const merged = mergeEvents(current, incoming, maxEvents);
      eventsRef.current = merged;
      return merged;
    });
  }, []);

  useEffect(() => {
    if (!runID) {
      setError("No simulation run was selected.");
      setLoading(false);
      return;
    }

    let mounted = true;
    let nextPoll: number | undefined;

    const poll = async () => {
      try {
        const loadedRun = await getLog(runID);
        const initialLoad = eventsRef.current.length === 0 && latestSequenceRef.current === 0;
        const loadedEvents = await getLogEvents(runID, initialLoad && loadedRun.summary.eventCount > 0
          ? { beforeSequence: loadedRun.summary.eventCount + 1, limit: eventPageSize }
          : { afterSequence: latestSequenceRef.current, limit: eventPageSize });
        if (!mounted) return;
        runRef.current = loadedRun;
        setRun(loadedRun);
        latestSequenceRef.current = Math.max(
          latestSequenceRef.current,
          loadedEvents.lastSequence,
        );
        updateEvents(
          loadedEvents.events,
          loadedRun.summary.status === "running" || loadedRun.summary.status === "paused"
            ? maxLiveEvents
            : undefined,
        );
        setSelectedObject((currentObject) => currentObject ?? loadedRun.objects[0] ?? null);
        setError(null);
        setLoading(false);
        if (loadedRun.summary.status === "running" || loadedRun.summary.status === "paused") {
          nextPoll = window.setTimeout(() => void poll(), 1000);
        }
      } catch {
        if (mounted) {
          setError((currentError) => currentError ?? "Unable to load this simulation log.");
          setLoading(false);
          nextPoll = window.setTimeout(() => void poll(), 1500);
        }
      }
    };

    void poll();
    return () => {
      mounted = false;
      if (nextPoll !== undefined) window.clearTimeout(nextPoll);
    };
  }, [getLog, getLogEvents, runID, updateEvents]);

  const loadOlderEvents = useCallback(async () => {
    if (!runID || loadingOlderEvents || eventsRef.current.length === 0) return;
    const oldestSequence = eventsRef.current[0].sequence;
    if (oldestSequence <= 1) return;
    setLoadingOlderEvents(true);
    try {
      const response = await getLogEvents(runID, {
        beforeSequence: oldestSequence,
        limit: eventPageSize,
      });
      const merged = mergeEvents(response.events, eventsRef.current);
      eventsRef.current = merged;
      setEvents(merged);
    } finally {
      setLoadingOlderEvents(false);
    }
  }, [getLogEvents, loadingOlderEvents, runID]);

  const filteredObjects = useMemo(() => {
    const normalized = objectFilter.trim().toLowerCase();
    if (!run || normalized.length === 0) return run?.objects ?? [];
    return run.objects.filter((object) =>
      `${object.name} ${object.identifier} ${object.kind}`.toLowerCase().includes(normalized),
    );
  }, [objectFilter, run]);

  const loadOlderObjectEvents = useCallback(async () => {
    if (!runID || !selectedObject || loadingOlderObjectEvents || objectEventsRef.current.length === 0) return;
    const oldestSequence = objectEventsRef.current[0].sequence;
    if (oldestSequence <= 1) return;
    setLoadingOlderObjectEvents(true);
    try {
      const filter = selectedObject.kind === "device"
        ? { deviceID: selectedObject.id }
        : { gatewayID: selectedObject.id };
      const response = await getLogEvents(runID, {
        ...filter,
        beforeSequence: oldestSequence,
        limit: eventPageSize,
      });
      const merged = mergeEvents(response.events, objectEventsRef.current);
      objectEventsRef.current = merged;
      setObjectEvents(merged);
    } finally {
      setLoadingOlderObjectEvents(false);
    }
  }, [getLogEvents, loadingOlderObjectEvents, runID, selectedObject]);

  useEffect(() => {
    if (!runID || !runRef.current || !selectedObject) return;
    let mounted = true;
    let nextPoll: number | undefined;
    const filter = selectedObject.kind === "device"
      ? { deviceID: selectedObject.id }
      : { gatewayID: selectedObject.id };

    const pollObjectEvents = async () => {
      try {
        const response = await getLogEvents(runID, {
          ...filter,
          afterSequence: objectLatestSequenceRef.current,
          beforeSequence: objectLatestSequenceRef.current === 0
            ? (runRef.current?.summary.eventCount ?? 0) + 1
            : undefined,
          limit: eventPageSize,
        });
        if (!mounted) return;
        const merged = mergeEvents(objectEventsRef.current, response.events);
        objectEventsRef.current = merged;
        objectLatestSequenceRef.current = Math.max(
          objectLatestSequenceRef.current,
          response.lastSequence,
        );
        setObjectEvents(merged);
        setObjectLoading(false);
        const status = runRef.current?.summary.status;
        if (status === "running" || status === "paused") {
          nextPoll = window.setTimeout(() => void pollObjectEvents(), 1000);
        }
      } catch {
        if (mounted) setObjectLoading(false);
      }
    };

    objectEventsRef.current = [];
    objectLatestSequenceRef.current = 0;
    setObjectEvents([]);
    setObjectLoading(true);
    void pollObjectEvents();
    return () => {
      mounted = false;
      if (nextPoll !== undefined) window.clearTimeout(nextPoll);
    };
  }, [getLogEvents, runID, selectedObject]);

  if (loading) return <div className="logs-page__loading"><Spinner /></div>;

  if (error || !run) {
    return <section className="logs-page"><p className="logs-page__empty">{error ?? "Simulation log not found."}</p></section>;
  }

  return (
    <section className="logs-page logs-page--detail">
      <header className="logs-page__header">
        <div>
          <span className="eyebrow">Simulation run</span>
          <h1>{formatDate(run.summary.startedAtMilliseconds)}</h1>
          <p>{run.summary.eventCount} recorded events · {run.summary.deviceCount} devices · {run.summary.gatewayCount} gateways</p>
        </div>
        <div className="logs-page__header-actions">
          <span className={`logs-page__status logs-page__status--${run.summary.status}`}>{formatEventType(run.summary.status)}</span>
          <Link href="/logs"><Button value="Back" /></Link>
        </div>
      </header>
      <div className="logs-page__detail-grid">
        <div className="logs-page__objects-column">
          <section className="logs-page__panel logs-page__panel--objects">
            <div className="logs-page__panel-head">
              <div><span className="eyebrow">Object history</span><h2>Devices and gateways</h2></div>
              <input
                aria-label="Search devices and gateways"
                className="logs-page__object-filter"
                onChange={(event) => setObjectFilter(event.target.value)}
                placeholder="Search objects"
                value={objectFilter}
              />
            </div>
            <VirtualObjectList objects={filteredObjects} onSelect={setSelectedObject} selectedObject={selectedObject} />
          </section>
          <section className="logs-page__panel logs-page__panel--object-events">
            <div className="logs-page__panel-head"><span className="eyebrow">Selected object</span><h2>{selectedObject?.name ?? "Object events"}</h2></div>
            {objectLoading ? <Spinner /> : (
              <>
                {loadingOlderObjectEvents ? <small className="logs-page__loading-label">Loading older events…</small> : null}
                <VirtualEventList
                  autoScroll={eventListNearEnd}
                  events={objectEvents}
                  emptyMessage="No events for this object."
                  onReachStart={() => void loadOlderObjectEvents()}
                />
              </>
            )}
          </section>
        </div>
        <section className="logs-page__panel logs-page__panel--events">
          <span className="eyebrow">Event stream</span>
          <h2>All recorded events</h2>
          {loadingOlderEvents ? <small className="logs-page__loading-label">Loading older events…</small> : null}
          <VirtualEventList
            autoScroll={eventListNearEnd}
            events={events}
            emptyMessage="No events were recorded."
            onNearEnd={setEventListNearEnd}
            onReachStart={() => void loadOlderEvents()}
          />
        </section>
      </div>
    </section>
  );
};

const LogDetailsPage = () => (
  <Suspense fallback={<div className="logs-page__loading"><Spinner /></div>}>
    <LogDetailsContent />
  </Suspense>
);

export default LogDetailsPage;
