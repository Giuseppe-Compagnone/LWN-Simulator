"use client";

import { SensorMapProps } from "./SensorMap.types";
import Map, { Layer, Marker, Source } from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";
import { Spinner, Theme, useThemeService } from "@lwn-simulator/ui-components";
import { useEffect, useState } from "react";

const DEFAULT_POSITION = {
  longitude: 12.4964,
  latitude: 41.9028,
  zoom: 12,
};

const SensorMap = (props: SensorMapProps) => {
  const themeService = useThemeService();

  const [position, setPosition] = useState(DEFAULT_POSITION);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!navigator.geolocation) {
      setLoading(false);
      return;
    }

    navigator.geolocation.getCurrentPosition(
      ({ coords }) => {
        setPosition({
          longitude: coords.longitude,
          latitude: coords.latitude,
          zoom: 12,
        });

        setLoading(false);
      },
      () => {
        setPosition(DEFAULT_POSITION);
        setLoading(false);
      },
      {
        enableHighAccuracy: true,
        timeout: 15000,
        maximumAge: 0,
      },
    );
  }, []);

  const markerA = {
    longitude: 9.19,
    latitude: 45.4642,
  };

  const markerB = {
    longitude: 9.19,
    latitude: 45.4742,
  };

  const lineGeoJSON = {
    type: "Feature" as const,
    geometry: {
      type: "LineString" as const,
      coordinates: [
        [markerA.longitude, markerA.latitude],
        [markerB.longitude, markerB.latitude],
      ],
    },
    properties: {},
  };

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
          themeService.theme === Theme.Dark
            ? "/dark-map-theme.json"
            : "/light-map-theme.json"
        }
        attributionControl={false}
        dragRotate={false}
        touchZoomRotate={false}
      >
        <Source id="marker-line" type="geojson" data={lineGeoJSON}>
          <Layer
            id="marker-line-layer"
            type="line"
            paint={{
              "line-color":
                themeService.theme === Theme.Dark ? "#38bdf8" : "#005cff",
              "line-width": 3,
              "line-opacity": 0.8,
              "line-dasharray": [2, 2],
            }}
          />
        </Source>
        <Marker
          longitude={markerA.longitude}
          latitude={markerA.latitude}
          anchor="center"
        >
          <span className="material-symbols-outlined marker sensor-marker">
            sensors
          </span>
        </Marker>
        <Marker
          longitude={markerB.longitude}
          latitude={markerB.latitude}
          anchor="center"
        >
          <span className="material-symbols-outlined marker gateway-marker">
            router
          </span>
        </Marker>
      </Map>
    </div>
  );
};

export default SensorMap;
