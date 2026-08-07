"use client";

import { SensorMapProps } from "./SensorMap.types";
import Map from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";
import { Spinner, Theme, useThemeService } from "@lwn-simulator/ui-components";
import { useEffect, useState } from "react";

const DEFAULT_POSITION = {
  longitude: 12.4964,
  latitude: 41.9028,
  zoom: 12,
};

const SensorMap = (props: SensorMapProps) => {
  const themeLogic = useThemeService();

  const [position, setPosition] = useState(DEFAULT_POSITION);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!navigator.geolocation) {
      setLoading(false);
      return;
    }

    navigator.geolocation.getCurrentPosition(
      ({ coords }) => {
        console.log("SUCCESS", coords.longitude, coords.latitude);

        setPosition({
          longitude: coords.longitude,
          latitude: coords.latitude,
          zoom: 12,
        });

        setLoading(false);
      },
      (error) => {
        console.log("GEO ERROR", error.code, error.message);

        setPosition(DEFAULT_POSITION);
        setLoading(false);
      },
      {
        enableHighAccuracy: false,
        timeout: 15000,
        maximumAge: 0,
      },
    );
  }, []);

  if (loading) {
    return (
      <div className="sensor-map">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="sensor-map">
      <Map
        initialViewState={position}
        style={{
          width: "100%",
          height: "100%",
        }}
        mapStyle={
          themeLogic.theme === Theme.Dark
            ? "/dark-map-theme.json"
            : "/light-map-theme.json"
        }
        attributionControl={false}
      />
    </div>
  );
};

export default SensorMap;
