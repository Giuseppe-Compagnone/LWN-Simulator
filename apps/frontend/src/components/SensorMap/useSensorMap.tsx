"use client";

import { MapLayerMouseEvent } from "maplibre-gl";
import {
  useSensorMapProps,
  SensorMapLogic,
  SensorMapMode,
  SensorMapPos,
} from "./SensorMap.types";
import { useEffect, useState } from "react";
import { NotificationHandler } from "@lwn-simulator/ui-components";

export const useSensorMap = (props: useSensorMapProps): SensorMapLogic => {
  const mode = props.mode || SensorMapMode.Sensor;

  // States
  const [markerPos, setMarkerPos] = useState<SensorMapPos | null>(null);
  const [selectedPos, setSelectedPos] = useState<SensorMapPos | null>(null);

  // Functions
  const getAltitude = async (lat: number, lng: number): Promise<number> => {
    let alt = 0;

    try {
      const response = await fetch(
        `https://api.open-elevation.com/api/v1/lookup?locations=${lat},${lng}`,
      );

      const data = await response.json();

      alt = data.results[0].elevation as number;
    } catch {
      NotificationHandler.instance.error("Failed to retrieve altitude");
    }

    return alt;
  };

  const handleClick = async (e: MapLayerMouseEvent) => {
    const { lat, lng } = e.lngLat;

    const alt = await getAltitude(lat, lng);

    setSelectedPos({
      lat,
      lng,
      alt,
    });
  };

  const updatePos = (lat: number, lng: number) => {
    if (markerPos && markerPos.lat == lat && markerPos.lng == lng) return;

    setMarkerPos({
      lat,
      lng,
      alt: markerPos?.alt || 0,
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
