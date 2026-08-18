"use client";

import { Layer, Source } from "react-map-gl/maplibre";
import { LinkMarkerProps } from "./LinkMarker.types";
import { Theme, useThemeService } from "@lwn-simulator/ui-components";

const LinkMarker = (props: LinkMarkerProps) => {
  // Hooks
  const themeService = useThemeService();
  return (
    <Source
      id="marker-line"
      type="geojson"
      data={{
        type: "Feature" as const,
        geometry: {
          type: "LineString" as const,
          coordinates: [
            [props.from.longitude, props.from.latitude],
            [props.to.longitude, props.to.latitude],
          ],
        },
        properties: {},
      }}
    >
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
  );
};

export default LinkMarker;
