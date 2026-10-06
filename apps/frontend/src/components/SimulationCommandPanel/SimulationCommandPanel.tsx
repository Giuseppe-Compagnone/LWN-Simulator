"use client";

import {
  SimulationMACCommand,
  SimulationMACCommandType,
} from "@lwn-simulator/contracts";
import {
  booleanCheckboxField,
  ButtonType,
  Card,
  CardLayout,
  Form,
  FormValue,
  NotificationHandler,
  selectField,
  textAreaField,
  textField,
} from "@lwn-simulator/ui-components";
import { useMemo, useState } from "react";
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
  info = `${label} value used by the LoRaWAN simulation.`,
): SimulationMACCommandFieldDescriptor => ({
  name,
  label,
  info,
  input: "number",
  minimum,
  maximum,
  required,
});

const booleanField = (
  name: SimulationMACCommandFieldDescriptor["name"],
  label: string,
  info = `${label} option applied to the simulated MAC command.`,
): SimulationMACCommandFieldDescriptor => ({ name, label, info, input: "checkbox" });

const MAC_COMMAND_FIELDS: SimulationMACCommandFields = {
  [SimulationMACCommandType.LinkCheckAns]: [
    numberField("margin", "Link margin", 0, 254),
    numberField("gatewayCount", "Gateway count", 0, 255),
  ],
  [SimulationMACCommandType.LinkAdrReq]: [
    numberField("dataRate", "Data rate", 0, 15),
    numberField("txPower", "TX power index", 0, 15),
    numberField("nbTrans", "Transmissions", 1, 15, true),
    numberField("channelMask", "Channel mask", 0, 65535),
    numberField("channelMaskControl", "Channel mask control", 0, 7),
  ],
  [SimulationMACCommandType.DutyCycleReq]: [
    numberField("maxDutyCycleExponent", "Duty-cycle exponent", 0, 15, true),
  ],
  [SimulationMACCommandType.RxParamSetupReq]: [
    numberField("dataRate", "RX2 data rate", 0, 15, true),
    numberField("rx1DataRateOffset", "RX1 data-rate offset", 0, 7),
    numberField("frequency", "RX2 frequency (Hz)", 1, 1677721500, true),
  ],
  [SimulationMACCommandType.DevStatusReq]: [],
  [SimulationMACCommandType.NewChannelReq]: [
    numberField("channelIndex", "Channel index", 0, 255, true),
    numberField("frequency", "Frequency (Hz)", 1, 1677721500, true),
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
    numberField("frequency", "Downlink frequency (Hz)", 1, 1677721500, true),
  ],
  [SimulationMACCommandType.DeviceTimeAns]: [
    numberField("deviceTimeMilliseconds", "Unix time (milliseconds)", 0, Number.MAX_SAFE_INTEGER, true),
  ],
  [SimulationMACCommandType.PingSlotInfoAns]: [
    numberField("pingSlotPeriodicity", "Ping-slot periodicity", 0, 7),
  ],
  [SimulationMACCommandType.PingSlotChannelReq]: [
    numberField("frequency", "Ping-slot frequency (Hz)", 1, 1677721500, true),
    numberField("dataRate", "Ping-slot data rate", 0, 15, true),
  ],
  [SimulationMACCommandType.BeaconFreqReq]: [
    numberField("frequency", "Beacon frequency (Hz)", 1, 1677721500, true),
  ],
};

const numericFormat = (value: string): string => value.replace(/[^0-9]/g, "");

const DATA_RATE_OPTIONS = [
  { value: "auto", displayed: <>Auto</> },
  ...Array.from({ length: 16 }, (_, dataRate) => ({
    value: String(dataRate),
    displayed: <>DR{dataRate}</>,
  })),
];

const UTF8_PAYLOAD_MAX_LENGTH = 242;
const BASE64_PAYLOAD_MAX_LENGTH = 324;

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
  const [macCommandType, setMACCommandType] = useState(
    SimulationMACCommandType.DevStatusReq,
  );
  const [isPayloadBase64, setIsPayloadBase64] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [lastActionID, setLastActionID] = useState<string | null>(null);

  const macFields = MAC_COMMAND_FIELDS[macCommandType];
  const deviceOptions = useMemo(
    () =>
      availableDevices.map((device) => ({
        value: device.id,
        displayed: <>{device.name} · {device.devEUI}</>,
      })),
    [availableDevices],
  );

  const setCommandMode = (nextMode: SimulationCommandMode) => {
    setMode(nextMode);
    setValidationError(null);
    setLastActionID(null);
  };

  const createMACCommand = (
    values: Record<string, FormValue>,
  ): SimulationMACCommand | null => {
    const commandValues: Record<string, number | boolean> = {};

    for (const field of macFields) {
      const rawValue = values[field.name];
      if (field.input === "checkbox") {
        commandValues[field.name] = Boolean(rawValue);
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
      commandValues[field.name] = value;
    }

    return { type: macCommandType, ...commandValues } as SimulationMACCommand;
  };

  const submit = async (values: Record<string, FormValue>) => {
    setValidationError(null);
    setLastActionID(null);
    const selectedDeviceID = String(values.deviceID ?? "");

    if (!selectedDeviceID) {
      setValidationError("Select an active device");
      return;
    }

    try {
      let response;
      if (mode === SimulationCommandMode.Uplink) {
        response = await props.onQueueUplink({ deviceID: selectedDeviceID });
      } else if (mode === SimulationCommandMode.Downlink) {
        const payload = String(values.payload ?? "");
        const fPort = Number(values.fPort);
        const dataRateValue = String(values.dataRate ?? "");
        const dataRate =
          dataRateValue === "" || dataRateValue === "auto"
            ? undefined
            : Number(dataRateValue);

        if (!payload.trim()) {
          setValidationError("Enter a downlink payload");
          return;
        }
        if (!Number.isInteger(fPort) || fPort < 1 || fPort > 223) {
          setValidationError("FPort must be an integer between 1 and 223");
          return;
        }
        if (
          dataRate !== undefined &&
          (!Number.isInteger(dataRate) || dataRate < 0 || dataRate > 15)
        ) {
          setValidationError("Data rate must be an integer between 0 and 15");
          return;
        }
        response = await props.onQueueDownlink({
          deviceID: selectedDeviceID,
          payload: isPayloadBase64 ? payload : encodeUTF8(payload),
          fPort,
          ...(dataRate === undefined ? {} : { dataRate }),
          confirmed: Boolean(values.confirmed),
          ack: Boolean(values.ack),
        });
      } else {
        const command = createMACCommand(values);
        if (!command) return;
        response = await props.onQueueMACCommand({
          deviceID: selectedDeviceID,
          command,
        });
      }

      setLastActionID(response.id);
      NotificationHandler.instance.success(`${formatEnum(mode)} queued successfully`);
    } catch (error) {
      const message =
        error instanceof Error ? error.message : `Failed to queue ${formatEnum(mode)}`;
      setValidationError(message);
      NotificationHandler.instance.error(message);
    }
  };

  const fields = useMemo(() => {
    const targetDevice = selectField({
      name: "deviceID",
      label: "Target device",
      value: availableDevices[0]?.id ?? null,
      error: null,
      options: deviceOptions,
      placeholder: availableDevices.length ? "Select an active device" : "No active devices",
      disabled: !props.enabled || availableDevices.length === 0,
      info: {
        default: "Select the active device that will receive or transmit the simulated traffic.",
      },
    });

    if (mode === SimulationCommandMode.Uplink) return [targetDevice];

    if (mode === SimulationCommandMode.Downlink) {
      return [
        targetDevice,
        textAreaField({
          name: "payload",
          label: `Payload (${isPayloadBase64 ? "Base64" : "UTF-8"})`,
          value: "",
          error: null,
          placeholder: "Message delivered to the selected device",
          charsMax: isPayloadBase64
            ? BASE64_PAYLOAD_MAX_LENGTH
            : UTF8_PAYLOAD_MAX_LENGTH,
          required: true,
          info: {
            default: isPayloadBase64
              ? "Base64 encoded application payload. The value is sent unchanged."
              : "UTF-8 application payload. It is encoded to bytes before transmission.",
          },
          validations: isPayloadBase64
            ? [
                {
                  rule: /^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$/,
                  error: "Payload must be valid Base64",
                },
              ]
            : undefined,
        }),
        booleanCheckboxField({
          name: "payloadBase64",
          label: "Base64 payload",
          text: "Payload is Base64 encoded",
          value: isPayloadBase64,
          error: null,
          info: {
            default:
              "Enable this when the payload already contains a Base64 representation of the bytes to send.",
          },
          onChange: (logic) => {
            setIsPayloadBase64(Boolean(logic.fieldsState.payloadBase64.value));
          },
        }),
        textField({
          name: "fPort",
          label: "FPort",
          value: "1",
          error: null,
          required: true,
          format: numericFormat,
          info: {
            default:
              "Application port. Valid LoRaWAN application ports range from 1 to 223.",
          },
          validations: [
            {
              rule: /^(?:[1-9]|[1-9]\d|1\d\d|2[0-1]\d|22[0-3])$/,
              error: "FPort must be an integer between 1 and 223",
            },
          ],
        }),
        selectField({
          name: "dataRate",
          label: "Data rate",
          value: "auto",
          error: null,
          options: DATA_RATE_OPTIONS,
          info: {
            default:
              "Select a fixed regional data rate or Auto to let the engine choose it.",
          },
        }),
        booleanCheckboxField({
          name: "confirmed",
          label: "Confirmed",
          text: "Confirmed downlink",
          value: false,
          error: null,
          info: {
            default:
              "Request an acknowledgement from the device for this downlink.",
          },
        }),
        booleanCheckboxField({
          name: "ack",
          label: "ACK",
          text: "Acknowledge confirmed uplink",
          value: false,
          error: null,
          info: {
            default:
              "Set the LoRaWAN ACK bit to acknowledge a confirmed uplink.",
          },
        }),
      ];
    }

    return [
      targetDevice,
      selectField({
        name: "macCommandType",
        label: "MAC command",
        value: macCommandType,
        error: null,
        info: {
          default:
            "Select the MAC command to queue for the target device.",
        },
        options: Object.values(SimulationMACCommandType).map((commandType) => ({
          value: commandType,
          displayed: <>{formatEnum(commandType)}</>,
        })),
        onChange: (logic) => {
          const value = logic.fieldsState.macCommandType.value;
          if (typeof value === "string") {
            setMACCommandType(value as SimulationMACCommandType);
            setValidationError(null);
          }
        },
      }),
      ...macFields.map((field) =>
        field.input === "checkbox"
          ? booleanCheckboxField({
              name: field.name,
              label: field.label,
              text: field.label,
              value: false,
              error: null,
              info: { default: field.info },
            })
          : textField({
              name: field.name,
              label: field.label,
              value: "",
              error: null,
              required: field.required,
              placeholder: field.required ? "Required" : "Optional",
              format: numericFormat,
              info: { default: field.info },
            }),
      ),
    ];
  }, [
    availableDevices,
    deviceOptions,
    isPayloadBase64,
    macCommandType,
    macFields,
    mode,
    props.enabled,
  ]);

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

      <div className={`simulation-command-panel__body is-${mode}`}>
        <Form
          fields={fields}
          onSubmit={submit}
          submitButton={{
            value: `Queue ${formatEnum(mode)}`,
            type: mode === SimulationCommandMode.Uplink ? ButtonType.Primary : ButtonType.Outlined,
            disabled: !props.enabled || availableDevices.length === 0,
          }}
        />

        {mode === SimulationCommandMode.Uplink && (
          <div className="simulation-command-panel__explanation">
            <span className="material-symbols-outlined">north_east</span>
            <div>
              <strong>Schedule one uplink now</strong>
              <p>OTAA devices that are not joined will send a join request first.</p>
            </div>
          </div>
        )}

        {mode === SimulationCommandMode.MACCommand && macFields.length === 0 && (
          <div className="simulation-command-panel__explanation">
            <span className="material-symbols-outlined">settings_input_antenna</span>
            <div>
              <strong>No additional parameters</strong>
              <p>This command can be queued immediately.</p>
            </div>
          </div>
        )}
      </div>

      <div className="simulation-command-panel__footer">
        <div aria-live="polite">
          {validationError && <span className="is-error">{validationError}</span>}
          {!validationError && lastActionID && <span className="is-success">Queued · {lastActionID}</span>}
        </div>
      </div>
    </Card>
  );
};

export default SimulationCommandPanel;
