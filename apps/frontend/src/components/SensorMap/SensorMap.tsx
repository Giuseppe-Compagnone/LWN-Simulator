"use client";

import { SensorMapProps } from "./SensorMap.types";
import Map, { Layer, Marker, Source } from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";
import { Spinner, Theme, useThemeService } from "@lwn-simulator/ui-components";
import { useEffect, useState } from "react";
import { DeviceMarker, GatewayMarker, LinkMarker } from "./components";

const DEFAULT_POSITION = {
  longitude: 12.4964,
  latitude: 41.9028,
  zoom: 10,
};

const SensorMap = (props: SensorMapProps) => {
  // States
  const [position, setPosition] = useState(DEFAULT_POSITION);

  // Hooks
  const themeService = useThemeService();
  const [loading, setLoading] = useState(true);

  // Functions
  const createCircle = (
    longitude: number,
    latitude: number,
    radiusMeters: number,
    points = 64,
  ) => {
    const coordinates: [number, number][] = [];

    const earthRadius = 6371008.8;

    const lat = (latitude * Math.PI) / 180;
    const lon = (longitude * Math.PI) / 180;
    const angularDistance = radiusMeters / earthRadius;

    for (let i = 0; i <= points; i++) {
      const bearing = (i / points) * 2 * Math.PI;

      const circleLat = Math.asin(
        Math.sin(lat) * Math.cos(angularDistance) +
          Math.cos(lat) * Math.sin(angularDistance) * Math.cos(bearing),
      );

      const circleLon =
        lon +
        Math.atan2(
          Math.sin(bearing) * Math.sin(angularDistance) * Math.cos(lat),
          Math.cos(angularDistance) - Math.sin(lat) * Math.sin(circleLat),
        );

      coordinates.push([
        (circleLon * 180) / Math.PI,
        (circleLat * 180) / Math.PI,
      ]);
    }

    return {
      type: "Feature" as const,
      geometry: {
        type: "Polygon" as const,
        coordinates: [coordinates],
      },
      properties: {},
    };
  };

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
          zoom: 10,
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
          <>
            <Marker
              longitude={props.logic.markerPos.lng}
              latitude={props.logic.markerPos.lat}
              anchor="bottom"
            >
              <span className="material-symbols-outlined pos-marker">
                location_on
              </span>
            </Marker>
            {props.logic.antennaRange && (
              <>
                <Source
                  id="coverage-circle"
                  type="geojson"
                  data={createCircle(
                    props.logic.markerPos.lng,
                    props.logic.markerPos.lat,
                    props.logic.antennaRange,
                  )}
                >
                  <Layer
                    id="coverage-circle-line"
                    type="line"
                    paint={{
                      "line-color": "#fff",
                      "line-width": 2,
                      "line-opacity": 0.8,
                      "line-dasharray": [2, 4],
                    }}
                  />
                </Source>
              </>
            )}
          </>
        )}
      </Map>
    </div>
  );
};

export default SensorMap;
