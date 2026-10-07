import { useCallback, useEffect, useMemo, useState } from "react";
import WebSocketContext from "./WebSocketContext";
import { WebSocketClient } from "./WebSocketClient";
import {
  WebSocketProviderProps,
  WebSocketMessageListener,
} from "./WebSocket.types";

const WebSocketProvider = (props: WebSocketProviderProps) => {
  const client = useMemo(() => new WebSocketClient(props.baseUrl), [props.baseUrl]);
  const [connectionState, setConnectionState] = useState(client.connectionState);
  const [error, setError] = useState<Error | null>(client.lastError);

  useEffect(() => {
    const removeStateListener = client.onStateChange(setConnectionState);
    const removeErrorListener = client.onError(setError);

    return () => {
      removeStateListener();
      removeErrorListener();
      client.dispose();
    };
  }, [client]);

  useEffect(() => {
    client.connect("/ws");
    return () => client.disconnect();
  }, [client]);

  const connect = useCallback((path: string) => client.connect(path), [client]);
  const disconnect = useCallback(() => client.disconnect(), [client]);
  const send = useCallback((payload: string) => client.send(payload), [client]);
  const sendJSON = useCallback((payload: unknown) => client.sendJSON(payload), [client]);
  const subscribe = useCallback(
    (listener: WebSocketMessageListener) => client.subscribe(listener),
    [client],
  );

  const value = useMemo(
    () => ({
      connect,
      disconnect,
      send,
      sendJSON,
      subscribe,
      connectionState,
      error,
    }),
    [connect, disconnect, send, sendJSON, subscribe, connectionState, error],
  );

  return (
    <WebSocketContext.Provider value={value}>
      {props.children}
    </WebSocketContext.Provider>
  );
};

export default WebSocketProvider;
