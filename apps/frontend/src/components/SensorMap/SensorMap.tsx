"use client";

import { SensorMapProps } from "./SensorMap.types";
import Map, { Marker } from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";
import { Spinner, Theme, useThemeService } from "@lwn-simulator/ui-components";
import { useEffect, useState } from "react";
import { DeviceMarker, GatewayMarker, LinkMarker } from "./components";

const DEFAULT_POSITION = {
  longitude: 12.4964,
  latitude: 41.9028,
  zoom: 12,
};

const SensorMap = (props: SensorMapProps) => {
  // States
  const [position, setPosition] = useState(DEFAULT_POSITION);

  // Hooks
  const themeService = useThemeService();
  const [loading, setLoading] = useState(true);

  // Effects
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
        onClick={props.logic.handleClick}
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
        <DeviceMarker marker={markerA} />
        <GatewayMarker marker={markerB} />
        <LinkMarker from={markerA} to={markerB} />
        {props.logic.markerPos && (
          <Marker
            longitude={props.logic.markerPos.lng}
            latitude={props.logic.markerPos.lat}
            anchor="bottom"
          >
            <span className="material-symbols-outlined">location_on</span>
          </Marker>
        )}
      </Map>
    </div>
  );
};

export default SensorMap;
