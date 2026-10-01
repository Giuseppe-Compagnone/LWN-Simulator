"use client";

import { SensorMapMode, SensorMapProps } from "./SensorMap.types";
import Map, { Layer, MapRef, Marker, Source } from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";
import { Theme, useThemeService } from "@lwn-simulator/ui-components";
import { useCallback, useEffect, useRef, useState } from "react";
import { DeviceMarker, GatewayMarker } from "./components";

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
  const mapRef = useRef<MapRef | null>(null);

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
    if (props.mode === SensorMapMode.Coords || !navigator.geolocation) {
      return;
    }

    navigator.geolocation.getCurrentPosition(
      ({ coords }) => {
        setPosition({
          longitude: coords.longitude,
          latitude: coords.latitude,
          zoom: 10,
        });
        mapRef.current?.flyTo({
          center: [coords.longitude, coords.latitude],
          zoom: 10,
        });
      },
      () => {},
      {
        enableHighAccuracy: true,
        timeout: 15000,
        maximumAge: 0,
      },
    );
  }, [props.mode]);

  const centerOnMarker = useCallback(() => {
    const marker = props.logic.markerPos;
    if (props.mode !== SensorMapMode.Coords || !marker || !mapRef.current) return;

    mapRef.current.flyTo({
      center: [marker.lng, marker.lat],
      zoom: 15,
      duration: 0,
    });
  }, [props.logic.markerPos, props.mode]);

  useEffect(() => {
    centerOnMarker();
  }, [centerOnMarker]);

  return (
    <div className="sensor-map">
      <Map
        ref={mapRef}
        onLoad={centerOnMarker}
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
        {props.devices?.map((device) => (
          <DeviceMarker
            key={device.id}
            marker={{
              latitude: device.locationConfig.latitude,
              longitude: device.locationConfig.longitude,
            }}
          />
        ))}
        {props.gateways?.map((gateway) => (
          <GatewayMarker
            key={gateway.id}
            marker={{ latitude: gateway.latitude, longitude: gateway.longitude }}
          />
        ))}
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
