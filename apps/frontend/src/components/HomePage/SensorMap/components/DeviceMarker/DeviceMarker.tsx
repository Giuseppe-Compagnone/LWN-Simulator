"use client";

import { Marker } from "react-map-gl/maplibre";
import { DeviceMarkerProps } from "./DeviceMarker.types";

const DeviceMarker = (props: DeviceMarkerProps) => {
  return (
    <Marker
      longitude={props.marker.longitude}
      latitude={props.marker.latitude}
      anchor="center"
    >
      <span className="material-symbols-outlined marker device-marker">
        sensors
      </span>
    </Marker>
  );
};

export default DeviceMarker;
