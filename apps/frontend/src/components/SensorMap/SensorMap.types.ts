import { MapLayerMouseEvent } from "maplibre-gl";

export enum SensorMapMode {
  Sensor = "sensor",
  Coords = "coords",
}

export interface Marker {
  latitude: number;
  longitude: number;
}

export interface SensorMapPos {
  lat: number;
  lng: number;
  alt: number;
}

export interface useSensorMapProps {
  mode?: SensorMapMode;
}

export interface SensorMapLogic {
  onClick?: (e: MapLayerMouseEvent) => void;
  selectedPos: SensorMapPos | null;
}
export interface SensorMapProps {
  logic: SensorMapLogic;
}
