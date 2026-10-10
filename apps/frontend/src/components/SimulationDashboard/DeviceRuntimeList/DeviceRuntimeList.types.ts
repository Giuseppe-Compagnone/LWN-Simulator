import { RuntimeDevice } from "@lwn-simulator/contracts";

/** Properties accepted by the runtime device list. */
export interface DeviceRuntimeListProps {
  /** Runtime device sessions displayed by the list. */
  devices: Array<RuntimeDevice>;
}
