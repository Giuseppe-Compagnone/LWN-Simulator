import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  SimulationConfig,
  SimulationEvent,
  SimulationSnapshot,
  SimulationWebSocketMessage,
  SimulationWebSocketMessageType,
} from "@lwn-simulator/contracts";
import SimulationServiceContext from "./SimulationServiceContext";
import {
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

  const updateSnapshot = useCallback((next: SimulationSnapshot) => {
    snapshotRef.current = next;
    setSnapshot(next);
  }, []);

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
          setEvents((previous) =>
            SimulationService.appendEvent(previous, update.event as SimulationEvent),
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
  }, [connectionState, service, updateSnapshot]);

  const start = useCallback(
    async (config: SimulationConfig) => {
      const next = await service.start(config);
      setEvents([]);
      updateSnapshot(next);
      connect();
      return next;
    },
    [connect, service, updateSnapshot],
  );

  const pause = useCallback(async () => {
    const next = await service.pause();
    updateSnapshot(next);
    return next;
  }, [service, updateSnapshot]);

  const resume = useCallback(async () => {
    const next = await service.resume();
    updateSnapshot(next);
    return next;
  }, [service, updateSnapshot]);

  const stop = useCallback(async () => {
    const next = await service.stop();
    updateSnapshot(next);
    disconnect();
    return next;
  }, [disconnect, service, updateSnapshot]);

  const getSnapshot = useCallback(async () => {
    const next = await service.getSnapshot();
    updateSnapshot(next);
    return next;
  }, [service, updateSnapshot]);

  useEffect(() => disconnect, [disconnect]);

  const value = useMemo(
    (): SimulationServiceContent => ({
      start,
      pause,
      resume,
      stop,
      getSnapshot,
      connect,
      disconnect,
      snapshot,
      events,
      connectionState,
      error,
    }),
    [
      start,
      pause,
      resume,
      stop,
      getSnapshot,
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
