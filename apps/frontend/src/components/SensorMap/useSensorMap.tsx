"use client";

import { MapLayerMouseEvent } from "maplibre-gl";
import {
  useSensorMapProps,
  SensorMapLogic,
  SensorMapMode,
  SensorMapPos,
} from "./SensorMap.types";
import { useState } from "react";

export const useSensorMap = (props: useSensorMapProps): SensorMapLogic => {
  const mode = props.mode || SensorMapMode.Sensor;

  // States
  const [selectedPos, setSelectedPos] = useState<SensorMapPos | null>(null);

  return {
    onClick:
      mode == SensorMapMode.Coords
        ? async (e: MapLayerMouseEvent) => {
            const { lng, lat } = e.lngLat;

            const response = await fetch(
              `https://api.open-elevation.com/api/v1/lookup?locations=${lat},${lng}`,
            );

            const data = await response.json();

            setSelectedPos({
              lat,
              lng,
              alt: data.results[0].elevation,
            });
          }
        : undefined,
    selectedPos,
  };
};
