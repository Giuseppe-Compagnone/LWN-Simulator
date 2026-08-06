"use client";

import { SensorMapProps } from "./SensorMap.types";
import Map from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";

const SensorMap = (props: SensorMapProps) => {
  return (
    <div className="sensor-map">
      <Map
        initialViewState={{
          longitude: 12.4964,
          latitude: 41.9028,
          zoom: 6,
        }}
        style={{
          width: "100%",
          height: "100%",
        }}
        mapStyle={true ? "/dark-map-theme.json" : "/dark-map-theme.json"}
      />
    </div>
  );
};

export default SensorMap;
