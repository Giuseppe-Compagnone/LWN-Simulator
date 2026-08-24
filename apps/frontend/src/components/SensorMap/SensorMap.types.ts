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
  handleClick?: (e: MapLayerMouseEvent) => void;
  selectedPos: SensorMapPos | null;
  markerPos: SensorMapPos | null;
  updatePos: (lat: number, lng: number) => void;
}
export interface SensorMapProps {
  logic: SensorMapLogic;
}
