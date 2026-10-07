import { PropsWithChildren } from "react";

export type WebSocketConnectionState =
  | "disconnected"
  | "connecting"
  | "connected"
  | "reconnecting";

export type WebSocketMessageListener = (message: string) => void;
export type WebSocketStateListener = (
  state: WebSocketConnectionState,
) => void;

export interface WebSocketContextValue {
  connect: (path: string) => void;
  disconnect: () => void;
  send: (payload: string) => void;
  sendJSON: (payload: unknown) => void;
  subscribe: (listener: WebSocketMessageListener) => () => void;
  connectionState: WebSocketConnectionState;
  error: Error | null;
}

export interface WebSocketProviderProps extends PropsWithChildren {
  baseUrl: string;
}
