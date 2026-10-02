import { Marker } from "../../SensorMap.types";

export interface GatewayMarkerProps {
  marker: Marker;
  onClick?: () => void;
}
