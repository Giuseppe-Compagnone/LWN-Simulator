import { Marker } from "../../SensorMap.types";

export interface LinkMarkerProps {
  id?: string;
  from: Marker;
  to: Marker;
}
