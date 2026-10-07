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
import SimulationServiceContext from "./SimulationServiceContext";
import {
  GetSimulationEventsOptions,
  SimulationConnectionState,
  SimulationServiceContent,
  SimulationServiceProviderProps,
} from "./SimulationService.types";
import { SimulationService } from "./SimulationService";

const SimulationServiceProvider = (props: SimulationServiceProviderProps) => {
  const [snapshot, setSnapshot] = useState<SimulationSnapshot | null>(null);
  const [events, setEvents] = useState<Array<SimulationEvent>>([]);
  const [connectionState, setConnectionState] =
    useState<SimulationConnectionState>("disconnected");
  const [error, setError] = useState<Error | null>(null);
  const service = useMemo(
    () => new SimulationService(props.baseUrl),
    [props.baseUrl],
  );
  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const shouldReconnectRef = useRef(false);
  const snapshotRef = useRef<SimulationSnapshot | null>(null);
  const eventsRef = useRef<Array<SimulationEvent>>([]);

  const updateSnapshot = useCallback((next: SimulationSnapshot) => {
    snapshotRef.current = next;
    setSnapshot(next);
  }, []);

  const updateEvents = useCallback((next: Array<SimulationEvent>) => {
    eventsRef.current = next;
    setEvents(next);
  }, []);

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

  const disconnect = useCallback(() => {
    shouldReconnectRef.current = false;
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }
    socketRef.current?.close();
    socketRef.current = null;
    setConnectionState("disconnected");
  }, []);

  const connect = useCallback(() => {
    if (socketRef.current || connectionState === "connecting") {
      return;
    }

    shouldReconnectRef.current = true;
    setConnectionState("connecting");
    const socket = new WebSocket(service.getWebSocketUrl());
    socketRef.current = socket;

    socket.onopen = () => {
      setError(null);
      setConnectionState("connected");
      void getEvents().catch((eventsError: unknown) => {
        setError(
          eventsError instanceof Error
            ? eventsError
            : new Error("Unable to recover simulation events"),
        );
      });
    };
    socket.onmessage = (message) => {
      try {
        const update = JSON.parse(message.data as string) as SimulationWebSocketMessage;
        if (update.type === SimulationWebSocketMessageType.Error) {
          setError(new Error(update.error ?? "Simulation websocket error"));
          return;
        }
        if (update.snapshot) {
          updateSnapshot(update.snapshot);
        }
        if (
          update.type === SimulationWebSocketMessageType.Event &&
          update.event
        ) {
          updateEvents(
            SimulationService.appendEvent(
              eventsRef.current,
              update.event as SimulationEvent,
            ),
          );
        }
      } catch (parseError) {
        setError(
          parseError instanceof Error
            ? parseError
            : new Error("Invalid simulation websocket message"),
        );
      }
    };
    socket.onerror = () => {
      setError(new Error("Simulation websocket connection failed"));
    };
    socket.onclose = () => {
      socketRef.current = null;
      setConnectionState("disconnected");
      const status = snapshotRef.current?.state.status;
      if (
        shouldReconnectRef.current &&
        status !== "stopped" &&
        status !== "failed"
      ) {
        reconnectTimerRef.current = setTimeout(connect, 1000);
      }
    };
  }, [connectionState, getEvents, service, updateEvents, updateSnapshot]);

  const start = useCallback(
    async (config: SimulationConfig) => {
      const next = await service.start(config);
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

  const setSpeed = useCallback(async (request: Parameters<SimulationService["setSpeed"]>[0]) => {
    const response = await service.setSpeed(request);
    updateSnapshot(response);
    setError(null);
    return response;
  }, [service, updateSnapshot]);

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
      connectionState,
      error,
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
