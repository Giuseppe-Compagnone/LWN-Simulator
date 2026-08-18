"use client";

import { Marker } from "react-map-gl/maplibre";
import { GatewayMarkerProps } from "./GatewayMarker.types";

const GatewayMarker = (props: GatewayMarkerProps) => {
  return (
    <Marker
      longitude={props.marker.longitude}
      latitude={props.marker.latitude}
      anchor="center"
    >
      <span className="material-symbols-outlined marker gateway-marker">
        router
      </span>
    </Marker>
  );
};

export default GatewayMarker;
