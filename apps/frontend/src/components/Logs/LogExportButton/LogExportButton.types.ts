/** Options for downloading simulation logs belonging to the current profile. */
export interface LogExportButtonProps {
  /** Run identifier to export as JSON. Omit to download all runs in a ZIP archive. */
  runID?: string;
  /** Prevents exporting before the history has loaded or when it is empty. */
  disabled?: boolean;
}
