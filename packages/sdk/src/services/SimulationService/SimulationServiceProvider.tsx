import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  SimulationActionResponse,
  SimulationConfig,
  SimulationDownlinkRequest,
  SimulationEvent,
  SimulationEventsResponse,
  SimulationMACCommandRequest,
  SimulationSnapshot,
  SimulationStatus,
  SimulationUplinkRequest,
  SimulationWebSocketMessage,
  SimulationWebSocketMessageType,
} from "@lwn-simulator/contracts";
import { useWebSocket } from "../../websocket";
import SimulationServiceContext from "./SimulationServiceContext";
import {
  GetSimulationEventsOptions,
  SimulationConnectionState,
  SimulationServiceContent,
  SimulationServiceProviderProps,
} from "./SimulationService.types";
import { SimulationService } from "./SimulationService";

const realtimeRenderIntervalMilliseconds = 100;
const simulationWebSocketPath = "/simulation/ws";

const SimulationServiceProvider = (props: SimulationServiceProviderProps) => {
  const [snapshot, setSnapshot] = useState<SimulationSnapshot | null>(null);
  const [events, setEvents] = useState<Array<SimulationEvent>>([]);
  const [error, setError] = useState<Error | null>(null);
  const service = useMemo(
    () => new SimulationService(props.baseUrl),
    [props.baseUrl],
  );
  const {
    connect: connectWebSocket,
    disconnect: disconnectWebSocket,
    subscribe: subscribeWebSocket,
    connectionState,
    error: websocketError,
  } = useWebSocket();
  const eventsRef = useRef<Array<SimulationEvent>>([]);
  const pendingSnapshotRef = useRef<SimulationSnapshot | null>(null);
  const pendingEventsRef = useRef<Array<SimulationEvent>>([]);
  const renderTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const updateSnapshot = useCallback((next: SimulationSnapshot) => {
    setSnapshot(next);
  }, []);

  const updateEvents = useCallback((next: Array<SimulationEvent>) => {
    eventsRef.current = next;
    setEvents(next);
  }, []);

  const flushRealtimeUpdates = useCallback(() => {
    renderTimerRef.current = null;

    if (pendingSnapshotRef.current) {
      updateSnapshot(pendingSnapshotRef.current);
      pendingSnapshotRef.current = null;
    }

    if (pendingEventsRef.current.length > 0) {
      updateEvents(
        SimulationService.mergeEvents(
          eventsRef.current,
          pendingEventsRef.current,
        ),
      );
      pendingEventsRef.current = [];
    }
  }, [updateEvents, updateSnapshot]);

  const scheduleRealtimeRender = useCallback(() => {
    if (renderTimerRef.current) return;

    renderTimerRef.current = setTimeout(
      flushRealtimeUpdates,
      realtimeRenderIntervalMilliseconds,
    );
  }, [flushRealtimeUpdates]);

  const getEvents = useCallback(
    async (
      options: GetSimulationEventsOptions = {},
    ): Promise<SimulationEventsResponse> => {
      const lastEvent = eventsRef.current.at(-1);
      const response = await service.getEvents({
        afterSequence: options.afterSequence ?? lastEvent?.sequence ?? 0,
        limit: options.limit ?? 100,
      });
      updateEvents(
        SimulationService.mergeEvents(eventsRef.current, response.events),
      );
      setError(null);
      return response;
    },
    [service, updateEvents],
  );

  const handleWebSocketMessage = useCallback(
    (rawMessage: string) => {
      try {
        const update = JSON.parse(rawMessage) as SimulationWebSocketMessage;
        if (update.type === SimulationWebSocketMessageType.Error) {
          setError(new Error(update.error ?? "Simulation websocket error"));
          return;
        }
        if (update.snapshot) {
          pendingSnapshotRef.current = update.snapshot;
        }
        if (
          update.type === SimulationWebSocketMessageType.Event &&
          update.event
        ) {
          pendingEventsRef.current.push(update.event);
        }
        scheduleRealtimeRender();
      } catch (parseError) {
        setError(
          parseError instanceof Error
            ? parseError
            : new Error("Invalid simulation websocket message"),
        );
      }
    },
    [scheduleRealtimeRender],
  );

  useEffect(
    () => subscribeWebSocket(handleWebSocketMessage),
    [handleWebSocketMessage, subscribeWebSocket],
  );

  useEffect(() => {
    if (connectionState !== "connected") return;

    void getEvents().catch((eventsError: unknown) => {
      setError(
        eventsError instanceof Error
          ? eventsError
          : new Error("Unable to recover simulation events"),
      );
    });
  }, [connectionState, getEvents]);

  const disconnect = useCallback(() => {
    disconnectWebSocket();
    if (renderTimerRef.current) {
      clearTimeout(renderTimerRef.current);
      flushRealtimeUpdates();
    }
  }, [disconnectWebSocket, flushRealtimeUpdates]);

  const connect = useCallback(
    () => connectWebSocket(simulationWebSocketPath),
    [connectWebSocket],
  );

  const start = useCallback(
    async (config: SimulationConfig) => {
      const next = await service.start(config);
      pendingSnapshotRef.current = null;
      pendingEventsRef.current = [];
      updateEvents([]);
      setError(null);
      updateSnapshot(next);
      connect();
      return next;
    },
    [connect, service, updateEvents, updateSnapshot],
  );

  const pause = useCallback(async () => {
    const next = await service.pause();
    setError(null);
    updateSnapshot(next);
    return next;
  }, [service, updateSnapshot]);

  const resume = useCallback(async () => {
    const next = await service.resume();
    setError(null);
    updateSnapshot(next);
    return next;
  }, [service, updateSnapshot]);

  const stop = useCallback(async () => {
    const next = await service.stop();
    setError(null);
    updateSnapshot(next);
    disconnect();
    return next;
  }, [disconnect, service, updateSnapshot]);

  const getSnapshot = useCallback(async () => {
    const next = await service.getSnapshot();
    updateSnapshot(next);
    setError(null);
    if (
      next.state.status === SimulationStatus.Running ||
      next.state.status === SimulationStatus.Paused
    ) {
      connect();
    }
    return next;
  }, [connect, service, updateSnapshot]);

  const queueUplink = useCallback(
    async (
      req: SimulationUplinkRequest,
    ): Promise<SimulationActionResponse> => {
      const response = await service.queueUplink(req);
      setError(null);
      return response;
    },
    [service],
  );

  const queueDownlink = useCallback(
    async (
      req: SimulationDownlinkRequest,
    ): Promise<SimulationActionResponse> => {
      const response = await service.queueDownlink(req);
      setError(null);
      return response;
    },
    [service],
  );

  const queueMACCommand = useCallback(
    async (
      req: SimulationMACCommandRequest,
    ): Promise<SimulationActionResponse> => {
      const response = await service.queueMACCommand(req);
      setError(null);
      return response;
    },
    [service],
  );

  const getMetrics = useCallback(async (): Promise<string> => {
    const response = await service.getMetrics();
    setError(null);
    return response;
  }, [service]);

  const getLogs = useCallback(() => service.getLogs(), [service]);
  const getLog = useCallback((id: string) => service.getLog(id), [service]);
  const getLogEvents = useCallback(
    (id: string, options: GetSimulationEventsOptions = {}) =>
      service.getLogEvents(id, options),
    [service],
  );

  const setSpeed = useCallback(
    async (request: Parameters<SimulationService["setSpeed"]>[0]) => {
      const response = await service.setSpeed(request);
      updateSnapshot(response);
      setError(null);
      return response;
    },
    [service, updateSnapshot],
  );

  useEffect(() => disconnect, [disconnect]);

  const value = useMemo(
    (): SimulationServiceContent => ({
      start,
      setSpeed,
      pause,
      resume,
      stop,
      getSnapshot,
      queueUplink,
      queueDownlink,
      queueMACCommand,
      getEvents,
      getMetrics,
      getLogs,
      getLog,
      getLogEvents,
      connect,
      disconnect,
      snapshot,
      events,
      connectionState: connectionState as SimulationConnectionState,
      error: error ?? websocketError,
    }),
    [
      start,
      setSpeed,
      pause,
      resume,
      stop,
      getSnapshot,
      queueUplink,
      queueDownlink,
      queueMACCommand,
      getEvents,
      getMetrics,
      getLogs,
      getLog,
      getLogEvents,
      connect,
      disconnect,
      snapshot,
      events,
      connectionState,
      websocketError,
      error,
    ],
  );

  return (
    <SimulationServiceContext.Provider value={value}>
      {props.children}
    </SimulationServiceContext.Provider>
  );
};

export default SimulationServiceProvider;
