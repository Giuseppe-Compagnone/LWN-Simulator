import { Device } from "@lwn-simulator/contracts";
import { FormValue } from "@lwn-simulator/ui-components";

/** Properties accepted by the device create and edit form. */
export interface DeviceFormProps {
  /** Device being edited, or null when creating a device. */
  device: Device | null;
  /** Receives validated form values for persistence. */
  onSubmit(values: Record<string, FormValue>): void | Promise<void>;
}

