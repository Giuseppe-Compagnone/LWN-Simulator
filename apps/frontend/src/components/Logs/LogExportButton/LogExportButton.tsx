"use client";

import { useCallback } from "react";
import { useSimulationService } from "@lwn-simulator/sdk";
import { Button, ButtonType, NotificationHandler } from "@lwn-simulator/ui-components";
import { LogExportButtonProps } from "./LogExportButton.types";

export const LogExportButton = (props: LogExportButtonProps) => {
  // Hooks
  const { getLogExportURL } = useSimulationService();

  // Callbacks
  const download = useCallback(async () => {
    try {
      const response = await fetch(getLogExportURL(props.runID));
      if (!response.ok) {
        throw new Error("Unable to export simulation logs. Please try again.");
      }
      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement("a");
      anchor.href = url;
      anchor.download = props.runID
        ? `simulation-${props.runID}.json`
        : "simulation-logs.zip";
      anchor.hidden = true;
      document.body.appendChild(anchor);
      anchor.click();
      window.setTimeout(() => {
        anchor.remove();
        URL.revokeObjectURL(url);
      }, 1000);
    } catch (error) {
      NotificationHandler.instance.error(
        error instanceof Error ? error.message : "Unable to export simulation logs.",
      );
    }
  }, [getLogExportURL, props.runID]);

  return (
    <Button
      className="log-export-button"
      value={props.runID ? "Export JSON" : "Export all (ZIP)"}
      type={ButtonType.Outlined}
      disabled={props.disabled}
      onClick={download}
    />
  );
};
