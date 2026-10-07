import { createContext } from "react";
import { WebSocketContextValue } from "./WebSocket.types";

const WebSocketContext = createContext<WebSocketContextValue | null>(null);

export default WebSocketContext;
