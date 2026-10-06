"use client";

import { Layer, Source } from "react-map-gl/maplibre";
import { LinkMarkerProps } from "./LinkMarker.types";
import { Theme, useThemeService } from "@lwn-simulator/ui-components";

const LinkMarker = (props: LinkMarkerProps) => {
  // Hooks
  const themeService = useThemeService();
  const markerID = props.id ?? "marker-line";
  return (
    <Source
      id={`marker-line-${markerID}`}
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
        id={`marker-line-glow-${markerID}`}
        type="line"
        paint={{
          "line-color":
            themeService.theme === Theme.Dark ? "#38bdf8" : "#005cff",
          "line-width": 8,
          "line-opacity": 0.3,
          "line-blur": 2,
        }}
      />
      <Layer
        id={`marker-line-layer-${markerID}`}
        type="line"
        paint={{
          "line-color":
            themeService.theme === Theme.Dark ? "#38bdf8" : "#005cff",
          "line-width": 4,
          "line-opacity": 1,
          "line-dasharray": [2, 1.5],
        }}
      />
    </Source>
  );
};

export default LinkMarker;
