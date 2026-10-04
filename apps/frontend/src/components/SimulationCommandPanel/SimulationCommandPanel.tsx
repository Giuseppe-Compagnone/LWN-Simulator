"use client";

import {
  SimulationMACCommand,
  SimulationMACCommandType,
} from "@lwn-simulator/contracts";
import {
  Button,
  ButtonType,
  Card,
  CardLayout,
  NotificationHandler,
} from "@lwn-simulator/ui-components";
import { useEffect, useMemo, useState } from "react";
import {
  SimulationCommandMode,
  SimulationCommandPanelProps,
  SimulationMACCommandFieldDescriptor,
  SimulationMACCommandFields,
} from "./SimulationCommandPanel.types";

const formatEnum = (value: string): string =>
  value.replace(/-/g, " ").replace(/\b\w/g, (letter) => letter.toUpperCase());

const numberField = (
  name: SimulationMACCommandFieldDescriptor["name"],
  label: string,
  minimum: number,
  maximum: number,
  required = false,
  step = 1,
): SimulationMACCommandFieldDescriptor => ({
  name,
  label,
  input: "number",
  minimum,
  maximum,
  required,
  step,
});

const booleanField = (
  name: SimulationMACCommandFieldDescriptor["name"],
  label: string,
): SimulationMACCommandFieldDescriptor => ({ name, label, input: "checkbox" });

const MAC_COMMAND_FIELDS: SimulationMACCommandFields = {
  [SimulationMACCommandType.LinkCheckAns]: [
    numberField("margin", "Link margin", 0, 254),
    numberField("gatewayCount", "Gateway count", 0, 255),
  ],
  [SimulationMACCommandType.LinkAdrReq]: [
    numberField("dataRate", "Data rate", 0, 15),
    numberField("txPower", "TX power index", 0, 15),
    numberField("nbTrans", "Transmissions", 1, 15),
    numberField("channelMask", "Channel mask", 0, 65535),
    numberField("channelMaskControl", "Channel mask control", 0, 7),
  ],
  [SimulationMACCommandType.DutyCycleReq]: [
    numberField("maxDutyCycleExponent", "Duty-cycle exponent", 0, 15, true),
  ],
  [SimulationMACCommandType.RxParamSetupReq]: [
    numberField("dataRate", "RX2 data rate", 0, 15, true),
    numberField("rx1DataRateOffset", "RX1 data-rate offset", 0, 7),
    numberField("frequency", "RX2 frequency (Hz)", 1, 1677721500, true, 100),
  ],
  [SimulationMACCommandType.DevStatusReq]: [],
  [SimulationMACCommandType.NewChannelReq]: [
    numberField("channelIndex", "Channel index", 0, 255, true),
    numberField("frequency", "Frequency (Hz)", 1, 1677721500, true, 100),
    numberField("minimumDataRate", "Minimum data rate", 0, 15, true),
    numberField("maximumDataRate", "Maximum data rate", 0, 15, true),
  ],
  [SimulationMACCommandType.RxTimingSetupReq]: [
    numberField("delaySeconds", "Receive delay (seconds)", 1, 15, true),
  ],
  [SimulationMACCommandType.TxParamSetupReq]: [
    booleanField("uplinkDwellTime", "Uplink dwell time"),
    booleanField("downlinkDwellTime", "Downlink dwell time"),
    numberField("maximumEIRP", "Maximum EIRP index", 0, 15),
  ],
  [SimulationMACCommandType.DlChannelReq]: [
    numberField("channelIndex", "Channel index", 0, 255, true),
    numberField("frequency", "Downlink frequency (Hz)", 1, 1677721500, true, 100),
  ],
  [SimulationMACCommandType.DeviceTimeAns]: [
    numberField(
      "deviceTimeMilliseconds",
      "Unix time (milliseconds)",
      0,
      Number.MAX_SAFE_INTEGER,
      true,
    ),
  ],
  [SimulationMACCommandType.PingSlotInfoAns]: [
    numberField("pingSlotPeriodicity", "Ping-slot periodicity", 0, 7),
  ],
  [SimulationMACCommandType.PingSlotChannelReq]: [
    numberField("frequency", "Ping-slot frequency (Hz)", 1, 1677721500, true, 100),
    numberField("dataRate", "Ping-slot data rate", 0, 15, true),
  ],
  [SimulationMACCommandType.BeaconFreqReq]: [
    numberField("frequency", "Beacon frequency (Hz)", 1, 1677721500, true, 100),
  ],
};

const encodeUTF8 = (value: string): string => {
  const bytes = new TextEncoder().encode(value);
  return window.btoa(String.fromCharCode(...bytes));
};

const SimulationCommandPanel = (props: SimulationCommandPanelProps) => {
  const availableDevices = useMemo(
    () => props.devices.filter((device) => device.active),
    [props.devices],
  );
  const [mode, setMode] = useState(SimulationCommandMode.Uplink);
  const [deviceID, setDeviceID] = useState(availableDevices[0]?.id ?? "");
  const [payload, setPayload] = useState("");
  const [fPort, setFPort] = useState("1");
  const [dataRate, setDataRate] = useState("");
  const [confirmed, setConfirmed] = useState(false);
  const [fPending, setFPending] = useState(false);
  const [ack, setACK] = useState(false);
  const [macCommandType, setMACCommandType] = useState(
    SimulationMACCommandType.DevStatusReq,
  );
  const [macValues, setMACValues] = useState<Record<string, string | boolean>>(
    {},
  );
  const [validationError, setValidationError] = useState<string | null>(null);
  const [lastActionID, setLastActionID] = useState<string | null>(null);

  useEffect(() => {
    if (!availableDevices.some((device) => device.id === deviceID)) {
      setDeviceID(availableDevices[0]?.id ?? "");
    }
  }, [availableDevices, deviceID]);

  const macFields = MAC_COMMAND_FIELDS[macCommandType];
  const canSubmit = props.enabled && Boolean(deviceID);

  const setCommandMode = (nextMode: SimulationCommandMode) => {
    setMode(nextMode);
    setValidationError(null);
    setLastActionID(null);
  };

  const createMACCommand = (): SimulationMACCommand | null => {
    const values: Record<string, number | boolean> = {};

    for (const field of macFields) {
      const rawValue = macValues[field.name];
      if (field.input === "checkbox") {
        values[field.name] = Boolean(rawValue);
        continue;
      }
      if (rawValue === undefined || rawValue === "") {
        if (field.required) {
          setValidationError(`${field.label} is required`);
          return null;
        }
        continue;
      }
      const value = Number(rawValue);
      if (
        !Number.isFinite(value) ||
        (field.minimum !== undefined && value < field.minimum) ||
        (field.maximum !== undefined && value > field.maximum)
      ) {
        setValidationError(`${field.label} is outside the supported range`);
        return null;
      }
      values[field.name] = value;
    }

    return { type: macCommandType, ...values } as SimulationMACCommand;
  };

  const submit = async () => {
    setValidationError(null);
    setLastActionID(null);

    if (!deviceID) {
      setValidationError("Select an active device");
      return;
    }

    try {
      let response;
      if (mode === SimulationCommandMode.Uplink) {
        response = await props.onQueueUplink({ deviceID });
      } else if (mode === SimulationCommandMode.Downlink) {
        if (!payload.trim()) {
          setValidationError("Enter a downlink payload");
          return;
        }
        const parsedFPort = Number(fPort);
        const parsedDataRate = dataRate === "" ? undefined : Number(dataRate);
        if (!Number.isInteger(parsedFPort) || parsedFPort < 1 || parsedFPort > 223) {
          setValidationError("FPort must be an integer between 1 and 223");
          return;
        }
        if (
          parsedDataRate !== undefined &&
          (!Number.isInteger(parsedDataRate) || parsedDataRate < 0 || parsedDataRate > 15)
        ) {
          setValidationError("Data rate must be an integer between 0 and 15");
          return;
        }
        response = await props.onQueueDownlink({
          deviceID,
          payload: encodeUTF8(payload),
          fPort: parsedFPort,
          ...(parsedDataRate === undefined ? {} : { dataRate: parsedDataRate }),
          confirmed,
          fPending,
          ack,
        });
      } else {
        const command = createMACCommand();
        if (!command) return;
        response = await props.onQueueMACCommand({ deviceID, command });
      }

      setLastActionID(response.id);
      NotificationHandler.instance.success(
        `${formatEnum(mode)} queued successfully`,
      );
    } catch (error) {
      const message =
        error instanceof Error ? error.message : `Failed to queue ${formatEnum(mode)}`;
      setValidationError(message);
      NotificationHandler.instance.error(message);
    }
  };

  return (
    <Card className="simulation-command-panel" layout={CardLayout.Padded}>
      <div className="simulation-command-panel__header">
        <div>
          <span className="simulation-dashboard__kicker">TRAFFIC CONTROL</span>
          <h2>Network command console</h2>
          <p>Inject protocol traffic without changing the device configuration.</p>
        </div>
        <span className={`simulation-command-panel__availability ${props.enabled ? "is-ready" : ""}`}>
          <span /> {props.enabled ? "Engine ready" : "Start or resume the simulation"}
        </span>
      </div>

      <div className="simulation-command-panel__modes" role="tablist" aria-label="Command type">
        {Object.values(SimulationCommandMode).map((commandMode) => (
          <button
            key={commandMode}
            type="button"
            role="tab"
            aria-selected={mode === commandMode}
            className={mode === commandMode ? "is-selected" : ""}
            onClick={() => setCommandMode(commandMode)}
          >
            {formatEnum(commandMode)}
          </button>
        ))}
      </div>

      <div className="simulation-command-panel__body">
        <label className="simulation-command-panel__field simulation-command-panel__device">
          <span>Target device</span>
          <select value={deviceID} onChange={(event) => setDeviceID(event.target.value)}>
            {availableDevices.length === 0 && <option value="">No active devices</option>}
            {availableDevices.map((device) => (
              <option value={device.id} key={device.id}>
                {device.name} · {device.devEUI}
              </option>
            ))}
          </select>
        </label>

        {mode === SimulationCommandMode.Uplink && (
          <div className="simulation-command-panel__explanation">
            <span className="material-symbols-outlined">north_east</span>
            <div>
              <strong>Schedule one uplink now</strong>
              <p>OTAA devices that are not joined will send a join request first.</p>
            </div>
          </div>
        )}

        {mode === SimulationCommandMode.Downlink && (
          <div className="simulation-command-panel__form-grid">
            <label className="simulation-command-panel__field simulation-command-panel__field--wide">
              <span>Payload (UTF-8)</span>
              <textarea
                value={payload}
                maxLength={242}
                placeholder="Message delivered to the selected device"
                onChange={(event) => setPayload(event.target.value)}
              />
            </label>
            <label className="simulation-command-panel__field">
              <span>FPort</span>
              <input type="number" min="1" max="223" value={fPort} onChange={(event) => setFPort(event.target.value)} />
            </label>
            <label className="simulation-command-panel__field">
              <span>Data rate <small>automatic when empty</small></span>
              <input type="number" min="0" max="15" value={dataRate} placeholder="Auto" onChange={(event) => setDataRate(event.target.value)} />
            </label>
            <div className="simulation-command-panel__checks">
              <label><input type="checkbox" checked={confirmed} onChange={(event) => setConfirmed(event.target.checked)} /> Confirmed</label>
              <label><input type="checkbox" checked={fPending} onChange={(event) => setFPending(event.target.checked)} /> FPending</label>
              <label><input type="checkbox" checked={ack} onChange={(event) => setACK(event.target.checked)} /> ACK</label>
            </div>
          </div>
        )}

        {mode === SimulationCommandMode.MACCommand && (
          <div className="simulation-command-panel__form-grid">
            <label className="simulation-command-panel__field simulation-command-panel__field--wide">
              <span>MAC command</span>
              <select
                value={macCommandType}
                onChange={(event) => {
                  setMACCommandType(event.target.value as SimulationMACCommandType);
                  setMACValues({});
                  setValidationError(null);
                }}
              >
                {Object.values(SimulationMACCommandType).map((commandType) => (
                  <option value={commandType} key={commandType}>
                    {formatEnum(commandType)}
                  </option>
                ))}
              </select>
            </label>
            {macFields.length === 0 && (
              <div className="simulation-command-panel__explanation simulation-command-panel__field--wide">
                <span className="material-symbols-outlined">settings_input_antenna</span>
                <div><strong>No additional parameters</strong><p>This command can be queued immediately.</p></div>
              </div>
            )}
            {macFields.map((field) =>
              field.input === "checkbox" ? (
                <label className="simulation-command-panel__boolean" key={field.name}>
                  <input
                    type="checkbox"
                    checked={Boolean(macValues[field.name])}
                    onChange={(event) =>
                      setMACValues((previous) => ({ ...previous, [field.name]: event.target.checked }))
                    }
                  />
                  <span>{field.label}</span>
                </label>
              ) : (
                <label className="simulation-command-panel__field" key={field.name}>
                  <span>{field.label}{field.required ? " *" : ""}</span>
                  <input
                    type="number"
                    min={field.minimum}
                    max={field.maximum}
                    step={field.step}
                    value={(macValues[field.name] as string | undefined) ?? ""}
                    placeholder={field.placeholder ?? (field.required ? "Required" : "Optional")}
                    onChange={(event) =>
                      setMACValues((previous) => ({ ...previous, [field.name]: event.target.value }))
                    }
                  />
                </label>
              ),
            )}
          </div>
        )}
      </div>

      <div className="simulation-command-panel__footer">
        <div aria-live="polite">
          {validationError && <span className="is-error">{validationError}</span>}
          {!validationError && lastActionID && <span className="is-success">Queued · {lastActionID}</span>}
        </div>
        <Button
          value={`Queue ${formatEnum(mode)}`}
          type={mode === SimulationCommandMode.Uplink ? ButtonType.Primary : ButtonType.Outlined}
          disabled={!canSubmit}
          onClick={submit}
        />
      </div>
    </Card>
  );
};

export default SimulationCommandPanel;

