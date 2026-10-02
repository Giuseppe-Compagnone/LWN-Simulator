"use client";

import { SensorMapMode, SensorMapProps, SensorMapEntity } from "./SensorMap.types";
import Map, { Layer, MapRef, Marker, Popup, Source } from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";
import { Theme, useOutsideAlerter, useThemeService } from "@lwn-simulator/ui-components";
import { useCallback, useEffect, useRef, useState } from "react";
import { DeviceMarker, GatewayMarker } from "./components";

const DEFAULT_POSITION = { longitude: 15.083, latitude: 37.5079, zoom: 12 };

const formatEnumValue = (value: string): string =>
  value.replace(/[-_]/g, " ");

const SensorMap = (props: SensorMapProps) => {
  const [selectedEntity, setSelectedEntity] = useState<SensorMapEntity | null>(null);
  const themeService = useThemeService();
  const mapRef = useRef<MapRef | null>(null);
  const tooltipRef = useRef<HTMLDivElement | null>(null);

  useOutsideAlerter({
    ref: tooltipRef,
    onClickOutside: () => setSelectedEntity(null),
  });

  const createCircle = (longitude: number, latitude: number, radiusMeters: number, points = 64) => {
    const coordinates: [number, number][] = [];
    const earthRadius = 6371008.8;
    const lat = (latitude * Math.PI) / 180;
    const lon = (longitude * Math.PI) / 180;
    const angularDistance = radiusMeters / earthRadius;

    for (let i = 0; i <= points; i += 1) {
      const bearing = (i / points) * 2 * Math.PI;
      const circleLat = Math.asin(Math.sin(lat) * Math.cos(angularDistance) + Math.cos(lat) * Math.sin(angularDistance) * Math.cos(bearing));
      const circleLon = lon + Math.atan2(Math.sin(bearing) * Math.sin(angularDistance) * Math.cos(lat), Math.cos(angularDistance) - Math.sin(lat) * Math.sin(circleLat));
      coordinates.push([(circleLon * 180) / Math.PI, (circleLat * 180) / Math.PI]);
    }

    return { type: "Feature" as const, geometry: { type: "Polygon" as const, coordinates: [coordinates] }, properties: {} };
  };

  const centerOnMarker = useCallback(() => {
    const marker = props.logic.markerPos;
    if (props.mode !== SensorMapMode.Coords || !marker || !mapRef.current) return;
    mapRef.current.flyTo({ center: [marker.lng, marker.lat], zoom: 15, duration: 0 });
  }, [props.logic.markerPos, props.mode]);

  useEffect(() => { centerOnMarker(); }, [centerOnMarker]);

  const selectedPosition = selectedEntity
    ? selectedEntity.kind === "device"
      ? { latitude: selectedEntity.entity.locationConfig.latitude, longitude: selectedEntity.entity.locationConfig.longitude }
      : { latitude: selectedEntity.entity.latitude, longitude: selectedEntity.entity.longitude }
    : null;

  return (
    <div className="sensor-map">
      <Map
        ref={mapRef}
        onLoad={centerOnMarker}
        onClick={props.logic.handleClick}
        initialViewState={DEFAULT_POSITION}
        style={{ width: "100%", height: "100%" }}
        mapStyle={themeService.theme === Theme.Dark ? "/dark-map-theme.json" : "/light-map-theme.json"}
        attributionControl={false}
        dragRotate={false}
        touchZoomRotate={false}
      >
        {props.devices?.map((device) => (
          <DeviceMarker
            key={device.id}
            marker={{ latitude: device.locationConfig.latitude, longitude: device.locationConfig.longitude }}
            onClick={() => setSelectedEntity({ kind: "device", entity: device })}
          />
        ))}
        {props.gateways?.map((gateway) => (
          <GatewayMarker
            key={gateway.id}
            marker={{ latitude: gateway.latitude, longitude: gateway.longitude }}
            onClick={() => setSelectedEntity({ kind: "gateway", entity: gateway })}
          />
        ))}

        {selectedEntity && selectedPosition && (
          <Popup
            longitude={selectedPosition.longitude}
            latitude={selectedPosition.latitude}
            anchor="bottom"
            maxWidth="min(21rem, calc(100vw - 2rem))"
            closeButton={false}
            closeOnClick={false}
            onClose={() => setSelectedEntity(null)}
            className="sensor-map-tooltip"
          >
            <div ref={tooltipRef} className="sensor-map-tooltip-content">
              <div className="sensor-map-tooltip-header">
                <span className="sensor-map-tooltip-kind">
                  {selectedEntity.kind === "device" ? "Sensors" : "Router"}
                </span>
                <div><strong>{selectedEntity.entity.name}</strong><span>{selectedEntity.kind === "device" ? "Device" : "Gateway"}</span></div>
                <button type="button" aria-label="Close tooltip" onClick={() => setSelectedEntity(null)}>×</button>
              </div>
              <div className="sensor-map-tooltip-details">
                <span>Status <strong className={selectedEntity.entity.active ? "is-active" : "is-inactive"}>{selectedEntity.entity.active ? "Active" : "Inactive"}</strong></span>
                {selectedEntity.kind === "device" ? (
                  <><span>DevEUI <strong>{selectedEntity.entity.devEUI}</strong></span><span>Class <strong>{formatEnumValue(selectedEntity.entity.class)}</strong></span><span>Region <strong>{selectedEntity.entity.locationConfig.region}</strong></span></>
                ) : (
                  <><span>Gateway EUI <strong>{selectedEntity.entity.gatewayEUI}</strong></span><span>Type <strong>{formatEnumValue(selectedEntity.entity.type)}</strong></span><span>MAC <strong>{selectedEntity.entity.macAddress}</strong></span></>
                )}
              </div>
            </div>
          </Popup>
        )}

        {selectedEntity?.kind === "device" && (
          <Source id="device-coverage-circle" type="geojson" data={createCircle(selectedPosition!.longitude, selectedPosition!.latitude, selectedEntity.entity.advancedConfig.antennaRange)}>
            <Layer id="device-coverage-circle-line" type="line" paint={{ "line-color": "#fff", "line-width": 2, "line-opacity": 0.9, "line-dasharray": [2, 4] }} />
          </Source>
        )}

        {props.logic.markerPos && (
          <>
            <Marker longitude={props.logic.markerPos.lng} latitude={props.logic.markerPos.lat} anchor="bottom"><span className="material-symbols-outlined pos-marker">location_on</span></Marker>
            {props.logic.antennaRange && <Source id="coverage-circle" type="geojson" data={createCircle(props.logic.markerPos.lng, props.logic.markerPos.lat, props.logic.antennaRange)}><Layer id="coverage-circle-line" type="line" paint={{ "line-color": "#fff", "line-width": 2, "line-opacity": 0.8, "line-dasharray": [2, 4] }} /></Source>}
          </>
        )}
      </Map>
    </div>
  );
};

export default SensorMap;
