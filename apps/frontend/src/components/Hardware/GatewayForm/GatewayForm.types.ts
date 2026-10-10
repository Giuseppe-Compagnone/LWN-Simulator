import { Gateway } from "@lwn-simulator/contracts";
import { FormValue } from "@lwn-simulator/ui-components";

/** Properties accepted by the gateway create and edit form. */
export interface GatewayFormProps {
  /** Gateway being edited, or `null` when creating a gateway. */
  gateway: Gateway | null;
  /** Receives validated form values for persistence. */
  onSubmit(values: Record<string, FormValue>): void | Promise<void>;
}
