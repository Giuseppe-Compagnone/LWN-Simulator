"use client";

import { MapLayerMouseEvent } from "maplibre-gl";
import {
  useSensorMapProps,
  SensorMapLogic,
  SensorMapMode,
  SensorMapPos,
} from "./SensorMap.types";
import { useEffect, useState } from "react";

export const useSensorMap = (props: useSensorMapProps): SensorMapLogic => {
  const mode = props.mode || SensorMapMode.Sensor;

  // States
  const [markerPos, setMarkerPos] = useState<SensorMapPos | null>(null);
  const [selectedPos, setSelectedPos] = useState<SensorMapPos | null>(null);

  // Functions
  const handleClick = (e: MapLayerMouseEvent) => {
    const { lat, lng } = e.lngLat;

    setSelectedPos({
      lat,
      lng,
    });
  };

  const updatePos = (lat: number, lng: number) => {
    if (markerPos && markerPos.lat == lat && markerPos.lng == lng) return;

    setMarkerPos({
      lat,
      lng,
    });
  };

  // Effects
  useEffect(() => {
    setMarkerPos(selectedPos);
  }, [selectedPos]);

  return {
    handleClick: mode == SensorMapMode.Coords ? handleClick : undefined,
    selectedPos,
    markerPos,
    updatePos,
  };
};
