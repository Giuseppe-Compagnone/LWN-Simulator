"use client";

import { SensorMap, SensorMapMode, useSensorMap } from "@/components";
import {
  estimateRegion,
  getAltitude,
  latValidation,
  lngValidation,
  randomDevEUI,
  rx1DataRateOffsetRange,
  rx2DataRateOptions,
  rx2Frequency,
} from "@/utils";
import {
  CreateDeviceRequest,
  DeviceActivation,
  DeviceClass,
  DeviceMType,
  DeviceRegion,
  OversizedPayloadBehavior,
} from "@lwn-simulator/contracts";
import { useDeviceService } from "@lwn-simulator/sdk";
import {
  booleanCheckboxField,
  Button,
  ButtonLayout,
  ButtonType,
  Form,
  FormField,
  FormLogic,
  FormValue,
  NotificationHandler,
  PageHeader,
  radioField,
  selectField,
  SelectOption,
  textField,
} from "@lwn-simulator/ui-components";
import { useEffect, useRef } from "react";

const NewDevicePage = () => {
  // Hooks
  const formLogicRef = useRef<FormLogic | null>(null);
  const mapLogic = useSensorMap({ mode: SensorMapMode.Coords });
  const deviceService = useDeviceService();

  // Functions
  const handleSubmit = async (values: Record<string, FormValue>) => {
    console.table(values);

    const req: CreateDeviceRequest = {
      devEUI: values.devEUI as string,
      name: values.name as string,
      activation: values.activation as DeviceActivation,
      class: values.class as DeviceClass,
      locationConfig: {
        latitude: Number(values.latitude),
        longitude: Number(values.longitude),
        altitude: Number(values.altitude),
        region: values.region as DeviceRegion,
      },
      RX1Config: {
        delay: Number(values.rx1Delay),
        duration: Number(values.rx1Duration),
        dataRateOffset: values.rx1DRO as number,
      },
      RX2Config: {
        delay: Number(values.rx2Delay),
        duration: Number(values.rx2Duration),
        channelFrequency: Number(values.rx2CF),
        dataRate: values.rx2DR as number,
        ACKTimeout: Number(values.rx2ACKT),
      },
      advancedConfig: {
        antennaRange: Number(values.antennaRange),
        ADREnabled: values.ADREnabled as boolean,
      },
      frameConfig: {
        fPort: Number(values.fPort),
        retransmission: Number(values.retransmission),
        disableFrameCounterValidation:
          values.disableFrameCounterValidation as boolean,
        FCntUp: values.fCntUp ? Number(values.fCntUp) : undefined,
        FCntDown: values.fCntDown ? Number(values.fCntDown) : undefined,
      },
      payloadConfig: {
        uplinkInterval: Number(values.uplinkInterval),
        oversizedPayloadBehavior:
          values.oversizedPayloadBehavior as OversizedPayloadBehavior,
        MType: values.MType as DeviceMType,
        payload: values.payload as string,
        base64Encoded: values.base64Encoded as boolean,
      },
      ...((values.activation as DeviceActivation) == DeviceActivation.Oota
        ? {
            OOTAConfig: {
              joinEUI: values.joinEUI as string,
              appKey: values.appKey as string,
            },
          }
        : {}),
      ...((values.activation as DeviceActivation) == DeviceActivation.Abp
        ? {
            ABPConfig: {
              devAddr: values.devAddr as string,
              nwkSKey: values.nwkSKey as string,
              appSKey: values.appSKey as string,
            },
          }
        : {}),
    };

    try {
      const res = await deviceService.createDevice(req);
      console.log("[CREATE DEVICE]", res);
    } catch {
      NotificationHandler.instance.error("Failed to create device");
    }
  };

  // Effects
  useEffect(() => {
    if (formLogicRef.current && mapLogic.selectedPos) {
      if (
        formLogicRef.current.fieldsState.latitude.value !=
        mapLogic.selectedPos.lat.toString()
      )
        formLogicRef.current.setValue(
          "latitude",
          mapLogic.selectedPos.lat.toString(),
        );

      if (
        formLogicRef.current.fieldsState.longitude.value !=
        mapLogic.selectedPos.lng.toString()
      )
        formLogicRef.current.setValue(
          "longitude",
          mapLogic.selectedPos.lng.toString(),
        );

      const region = estimateRegion(
        mapLogic.selectedPos.lat,
        mapLogic.selectedPos.lng,
      );

      if (region) {
        formLogicRef.current.setValue("region", region);
      }
    }
  }, [mapLogic.selectedPos]);

  return (
    <div className="page new-device-page">
      <PageHeader title={"Create new Device"} />
      <div className="page-content">
        <div className="form-wrapper">
          <Form
            onSubmit={(values: Record<string, FormValue>) => {
              handleSubmit(values);
            }}
            submitButton={{
              value: "Create",
              className: "submit-button",
            }}
            onLogicReady={(logic) => {
              formLogicRef.current = logic;
            }}
            fields={[
              textField({
                name: "name",
                label: "Name",
                value: "",
                placeholder: "Device Name",
                error: null,
                required: true,
              }),
              textField({
                name: "devEUI",
                label: "DevEUI",
                value: "",
                placeholder: "70B3D57ED0001234",
                format: (raw: string) => {
                  return raw
                    .replace(/[^0-9A-Fa-f]/g, "")
                    .slice(0, 16)
                    .toUpperCase();
                },
                info: {
                  default:
                    "Unique identifier assigned to the LoRaWAN end device",
                },
                validations: [
                  {
                    rule: /^[0-9A-F]{16}$/,
                    error: "Invalid DevEUI",
                  },
                ],
                error: null,
                required: true,
                toolbar: (
                  <Button
                    value={
                      <span className="material-symbols-outlined">cached</span>
                    }
                    layout={ButtonLayout.Icon}
                    type={ButtonType.Outlined}
                    onClick={() => {
                      formLogicRef.current?.setValue("devEUI", randomDevEUI());
                    }}
                  />
                ),
              }),
              textField({
                name: "latitude",
                label: "Latitude",
                placeholder: "45.4642",
                value: "",
                error: null,
                info: { default: "Decimal degrees (-90 to 90)" },
                format: (raw: string) => {
                  return raw
                    .replace(/[^0-9.-]/g, "")
                    .replace(/(?!^)-/g, "")
                    .replace(/(\..*)\./g, "$1");
                },
                validations: [
                  {
                    rule: latValidation,
                    error: "Invalid latitude",
                  },
                ],
                required: true,
                onChange: (logic: FormLogic) => {
                  if (
                    logic.fieldsState.latitude.value &&
                    (logic.fieldsState.latitude.value as string).match(
                      latValidation,
                    ) &&
                    logic.fieldsState.longitude.value &&
                    (logic.fieldsState.longitude.value as string).match(
                      lngValidation,
                    )
                  ) {
                    mapLogic.updatePos(
                      Number(logic.fieldsState.latitude.value),
                      Number(logic.fieldsState.longitude.value),
                    );
                  }
                },
              }),
              textField({
                name: "longitude",
                label: "Longitude",
                placeholder: "9.1900",
                value: "",
                error: null,
                info: { default: "Decimal degrees (-180 to 180)" },
                format: (raw: string) => {
                  return raw
                    .replace(/[^0-9.-]/g, "")
                    .replace(/(?!^)-/g, "")
                    .replace(/(\..*)\./g, "$1");
                },
                validations: [
                  {
                    rule: lngValidation,
                    error: "Invalid longitude",
                  },
                ],
                required: true,
                onChange: (logic: FormLogic) => {
                  if (
                    logic.fieldsState.latitude.value &&
                    (logic.fieldsState.latitude.value as string).match(
                      latValidation,
                    ) &&
                    logic.fieldsState.longitude.value &&
                    (logic.fieldsState.longitude.value as string).match(
                      lngValidation,
                    )
                  ) {
                    mapLogic.updatePos(
                      Number(logic.fieldsState.latitude.value),
                      Number(logic.fieldsState.longitude.value),
                    );
                  }
                },
              }),
              textField({
                name: "altitude",
                label: "Altitude",
                placeholder: "120",
                value: "",
                error: null,
                info: { default: "Meters above sea level" },
                format: (raw: string) => {
                  return raw
                    .replace(/[^0-9.-]/g, "")
                    .replace(/(?!^)-/g, "")
                    .replace(/(\..*)\./g, "$1");
                },
                validations: [
                  {
                    rule: /^-?\d+(?:\.\d+)?$/,
                    error: "Invalid altitude",
                  },
                ],
                required: true,
                toolbar: (
                  <Button
                    value={
                      <span className="material-symbols-outlined">cached</span>
                    }
                    layout={ButtonLayout.Icon}
                    type={ButtonType.Outlined}
                    onClick={async () => {
                      if (
                        !formLogicRef.current ||
                        !(
                          formLogicRef.current.fieldsState.longitude
                            .value as string
                        ).match(lngValidation) ||
                        !(
                          formLogicRef.current.fieldsState.latitude
                            .value as string
                        ).match(latValidation)
                      ) {
                        NotificationHandler.instance.error(
                          "Select valid latitude and longitude",
                        );
                        return;
                      }
                      const res = await getAltitude(
                        Number(formLogicRef.current.fieldsState.latitude.value),
                        Number(
                          formLogicRef.current.fieldsState.longitude.value,
                        ),
                      );

                      formLogicRef.current.setValue("altitude", res.toString());
                    }}
                  />
                ),
              }),
              selectField({
                value: null,
                name: "region",
                label: "Region",
                error: null,
                placeholder: "Select a region",
                options: Object.values(DeviceRegion).map((reg: string) => {
                  return {
                    value: reg,
                  };
                }),
                info: { default: "LoRaWAN regional frequency plan" },
                onChange: (logic: FormLogic) => {
                  logic.setValue("rx1DRO", null);

                  if (logic.fieldsState.region.value) {
                    const range =
                      rx1DataRateOffsetRange[
                        logic.fieldsState.region.value as DeviceRegion
                      ];

                    const values: Array<SelectOption> = [];

                    for (let i = range[0]; i <= range[1]; i++) {
                      values.push({
                        value: i.toString(),
                      });
                    }

                    logic.setProp("rx1DRO", "options", values);
                    logic.setProp(
                      "rx2DR",
                      "options",
                      rx2DataRateOptions[
                        logic.fieldsState.region.value as DeviceRegion
                      ].map((opt) => {
                        return { value: opt };
                      }),
                    );

                    const freq =
                      rx2Frequency[
                        logic.fieldsState.region.value as DeviceRegion
                      ];

                    logic.setProp("rx2CF", "placeholder", freq.toString());
                    logic.setValue("rx2CF", freq.toString());
                  } else {
                    logic.setProp(
                      "rx1DRO",
                      "options",
                      [] as Array<SelectOption>,
                    );
                    logic.setProp(
                      "rx2DR",
                      "options",
                      [] as Array<SelectOption>,
                    );
                  }
                },
                required: true,
                toolbar: (
                  <Button
                    value={
                      <span className="material-symbols-outlined">cached</span>
                    }
                    layout={ButtonLayout.Icon}
                    type={ButtonType.Outlined}
                    onClick={async () => {
                      if (
                        !formLogicRef.current ||
                        !(
                          formLogicRef.current.fieldsState.longitude
                            .value as string
                        ).match(lngValidation) ||
                        !(
                          formLogicRef.current.fieldsState.latitude
                            .value as string
                        ).match(latValidation)
                      ) {
                        NotificationHandler.instance.error(
                          "Select valid latitude and longitude",
                        );
                        return;
                      }

                      const res = estimateRegion(
                        Number(formLogicRef.current.fieldsState.latitude.value),
                        Number(
                          formLogicRef.current.fieldsState.longitude.value,
                        ),
                      );

                      if (!res)
                        NotificationHandler.instance.error(
                          "Unable to estimate region",
                        );

                      formLogicRef.current.setValue("region", res);
                    }}
                  />
                ),
              }),
              radioField({
                value: null,
                name: "class",
                label: "Class",
                error: null,
                options: Object.keys(DeviceClass).map((key: string) => {
                  return {
                    value: DeviceClass[key as keyof typeof DeviceClass],
                    displayed: <>{key.replace("Class", "")}</>,
                  };
                }),
                info: {
                  default: "Device class",
                  [DeviceClass.ClassA]:
                    "Lowest power consumption, with downlink windows after each uplink",
                  [DeviceClass.ClassB]:
                    "Scheduled downlink windows, offering predictable network communication",
                  [DeviceClass.ClassC]:
                    "Nearly continuous downlink availability, with highest power consumption",
                },
                required: true,
              }),
              radioField({
                value: null,
                name: "activation",
                label: "Activation Type",
                error: null,
                options: Object.keys(DeviceActivation).map((key: string) => {
                  return {
                    value:
                      DeviceActivation[key as keyof typeof DeviceActivation],
                    displayed: <>{key.toUpperCase()}</>,
                  };
                }),
                info: {
                  default: "Device Activation Type",
                  [DeviceActivation.Oota]:
                    "Secure activation through network join using unique session credentials",
                  [DeviceActivation.Abp]:
                    "Direct activation using preconfigured device and session credentials",
                },
                required: true,
              }),
              textField({
                name: "joinEUI",
                label: "JoinEUI",
                value: "",
                placeholder: "0004A30B001C0530",
                format: (raw: string) => {
                  return raw
                    .replace(/[^0-9A-Fa-f]/g, "")
                    .slice(0, 16)
                    .toUpperCase();
                },
                validations: [
                  {
                    rule: /^[0-9A-F]{16}$/,
                    error: "Invalid JoinEUI",
                  },
                ],
                info: {
                  default:
                    "Identifies the application entity handling the device join",
                },
                error: null,
                required: true,
                display: (fieldsState: Record<string, FormField>) =>
                  fieldsState.activation.value == DeviceActivation.Oota,
              }),
              textField({
                name: "appKey",
                label: "AppKey",
                value: "",
                placeholder: "2B7E151628AED2A6ABF7158809CF4F3C",
                format: (raw: string) => {
                  return raw
                    .replace(/[^0-9A-Fa-f]/g, "")
                    .slice(0, 32)
                    .toUpperCase();
                },
                validations: [
                  {
                    rule: /^[0-9A-F]{32}$/,
                    error: "Invalid AppKey",
                  },
                ],
                info: {
                  default:
                    "Secret key used to securely authenticate an OTAA join",
                },
                error: null,
                required: true,
                display: (fieldsState: Record<string, FormField>) =>
                  fieldsState.activation.value == DeviceActivation.Oota,
              }),
              textField({
                name: "devAddr",
                label: "DevAddr",
                value: "",
                placeholder: "26011BDA",
                format: (raw: string) => {
                  return raw
                    .replace(/[^0-9A-Fa-f]/g, "")
                    .slice(0, 8)
                    .toUpperCase();
                },
                validations: [
                  {
                    rule: /^[0-9A-F]{8}$/,
                    error: "Invalid DevAddr",
                  },
                ],
                info: {
                  default:
                    "32-bit address assigned to the device during ABP activation",
                },
                error: null,
                required: true,
                display: (fieldsState: Record<string, FormField>) =>
                  fieldsState.activation.value == DeviceActivation.Abp,
              }),
              textField({
                name: "nwkSKey",
                label: "NwkSKey",
                value: "",
                placeholder: "00112233445566778899AABBCCDDEEFF",
                format: (raw: string) => {
                  return raw
                    .replace(/[^0-9A-Fa-f]/g, "")
                    .slice(0, 32)
                    .toUpperCase();
                },
                validations: [
                  {
                    rule: /^[0-9A-F]{32}$/,
                    error: "Invalid NwkSKey",
                  },
                ],
                info: {
                  default:
                    "Session key used to secure LoRaWAN network-layer communication",
                },
                error: null,
                required: true,
                display: (fieldsState: Record<string, FormField>) =>
                  fieldsState.activation.value == DeviceActivation.Abp,
              }),
              textField({
                name: "appSKey",
                label: "AppSKey",
                value: "",
                placeholder: "0004AFFEEDDCCBBAA9988776655443322110030B001C0530",
                format: (raw: string) => {
                  return raw
                    .replace(/[^0-9A-Fa-f]/g, "")
                    .slice(0, 32)
                    .toUpperCase();
                },
                validations: [
                  {
                    rule: /^[0-9A-F]{32}$/,
                    error: "Invalid AppSKey",
                  },
                ],
                info: {
                  default:
                    "Session key used to encrypt application payload data",
                },
                error: null,
                required: true,
                display: (fieldsState: Record<string, FormField>) =>
                  fieldsState.activation.value == DeviceActivation.Abp,
              }),
              textField({
                name: "rx1Delay",
                label: "RX1 Delay",
                value: "",
                placeholder: "1000",
                format: (raw: string) => {
                  return raw.replace(/[^0-9]/g, "");
                },
                info: {
                  default:
                    "Delay before opening the RX1 downlink reception window in milliseconds",
                },
                error: null,
                required: true,
              }),
              textField({
                name: "rx1Duration",
                label: "RX1 Duration",
                value: "",
                placeholder: "3000",
                format: (raw: string) => {
                  return raw.replace(/[^0-9]/g, "");
                },
                info: {
                  default:
                    "Duration of the RX1 downlink reception window in milliseconds",
                },
                error: null,
                required: true,
              }),
              selectField({
                value: null,
                name: "rx1DRO",
                label: "RX1 Data Rate offset",
                error: null,
                placeholder: "Select DRO",
                options: [],
                info: {
                  default: "Offset applied to the uplink data rate for RX1",
                },
                required: true,
                display: (fieldsState: Record<string, FormField>) =>
                  !!fieldsState.region.value,
              }),
              textField({
                name: "rx2Delay",
                label: "RX2 Delay",
                value: "",
                placeholder: "1000",
                format: (raw: string) => {
                  return raw.replace(/[^0-9]/g, "");
                },
                info: {
                  default:
                    "Delay before opening the RX2 downlink reception window in milliseconds",
                },
                error: null,
                required: true,
              }),
              textField({
                name: "rx2Duration",
                label: "RX2 Duration",
                value: "",
                placeholder: "3000",
                format: (raw: string) => {
                  return raw.replace(/[^0-9]/g, "");
                },
                info: {
                  default:
                    "Duration of the RX2 downlink reception window in milliseconds",
                },
                error: null,
                required: true,
              }),
              textField({
                name: "rx2CF",
                label: "RX2 Channel Frequency",
                value: "",
                placeholder: "",
                format: (raw: string) => {
                  return raw.replace(/[^0-9]/g, "");
                },
                info: {
                  default:
                    "Frequency used for receiving downlink messages during RX2, in Hz",
                },
                error: null,
                required: true,
                display: (fieldsState: Record<string, FormField>) =>
                  !!fieldsState.region.value,
              }),
              selectField({
                name: "rx2DR",
                label: "RX2 Data Rate",
                value: "",
                placeholder: "0",
                options: [],
                info: {
                  default:
                    "Data rate used for receiving downlink messages during RX2",
                },
                error: null,
                required: true,
                display: (fieldsState: Record<string, FormField>) =>
                  !!fieldsState.region.value,
              }),
              textField({
                name: "rx2ACKT",
                label: "RX2 ACK Timeout",
                value: "",
                placeholder: "2",
                format: (raw: string) => {
                  return raw.replace(/[^0-9.]/g, "").replace(/(\..*)\./g, "$1");
                },
                info: {
                  default:
                    "Time the device waits for an acknowledgement after a confirmed uplink",
                },
                error: null,
                required: true,
              }),
              textField({
                name: "fPort",
                label: "FPort",
                value: "",
                placeholder: "1",
                format: (raw: string) => {
                  return raw.replace(/[^0-9]/g, "");
                },
                info: {
                  default: "Application port used to identify the payload type",
                },
                validations: [
                  {
                    rule: /^(?:[1-9]|[1-9][0-9]|1[0-9]{2}|2[0-2][0-9]|23[0-3])$/,
                    error: "Fport must be between 1 and 223",
                  },
                ],
                error: null,
                required: true,
              }),
              textField({
                name: "retransmission",
                label: "Retransmission",
                value: "",
                placeholder: "0",
                format: (raw: string) => {
                  return raw.replace(/[^0-9]/g, "");
                },
                info: {
                  default:
                    "Number of times an uplink is retransmitted after transmission failure",
                },
                error: null,
                required: true,
              }),
              textField({
                name: "fCntUp",
                label: "FCntUp",
                value: "",
                placeholder: "0",
                format: (raw: string) => {
                  return raw.replace(/[^0-9]/g, "");
                },
                info: {
                  default:
                    "Uplink frame counter used to track transmitted frames",
                },
                error: null,
              }),
              textField({
                name: "fCntDown",
                label: "FCntDown",
                value: "",
                placeholder: "0",
                format: (raw: string) => {
                  return raw.replace(/[^0-9]/g, "");
                },
                info: {
                  default:
                    "Downlink frame counter used to track received frames",
                },
                error: null,
              }),
              booleanCheckboxField({
                value: false,
                name: "disableFrameCounterValidation",
                label: "Frame Counter Validation",
                error: null,
                text: "Disable",
                info: {
                  default: "Disables validation of downlink frame counters",
                },
              }),
              textField({
                name: "uplinkInterval",
                label: "Uplink Interval",
                value: "",
                placeholder: "10",
                format: (raw: string) => {
                  return raw.replace(/[^0-9]/g, "");
                },
                validations: [
                  {
                    rule: /^[1-9][0-9]*$/,
                    error: "Uplink Interval must be greater than 0",
                  },
                ],
                info: {
                  default:
                    "Time between consecutive uplink transmissions, in seconds",
                },
                error: null,
                required: true,
              }),
              radioField({
                value: null,
                name: "oversizedPayloadBehavior",
                label: "Oversized Payload Behavior",
                error: null,
                options: Object.keys(OversizedPayloadBehavior).map(
                  (key: string) => {
                    return {
                      value:
                        OversizedPayloadBehavior[
                          key as keyof typeof OversizedPayloadBehavior
                        ],
                      displayed: <>{key}</>,
                    };
                  },
                ),
                info: {
                  default:
                    "Defines how payloads exceeding the maximum size are handled",
                  [OversizedPayloadBehavior.Fragment]:
                    "Splits oversized payloads into multiple smaller transmissions",
                  [OversizedPayloadBehavior.Truncate]:
                    "Truncates oversized payloads to the maximum allowed size",
                },
                required: true,
              }),
              radioField({
                value: null,
                name: "MType",
                label: "MType",
                error: null,
                options: Object.keys(DeviceMType).map((key: string) => {
                  return {
                    value: DeviceMType[key as keyof typeof DeviceMType],
                    displayed: (
                      <>
                        {DeviceMType[
                          key as keyof typeof DeviceMType
                        ].replaceAll("-", " ")}
                      </>
                    ),
                  };
                }),
                info: {
                  default:
                    "Defines the LoRaWAN message type used for uplink transmissions",
                  [DeviceMType.ConfirmedDataUp]:
                    "Requires a downlink acknowledgement from the network",
                  [DeviceMType.UnconfirmedDataUp]:
                    "Does not require a downlink acknowledgement from the network",
                },
                required: true,
              }),
              textField({
                name: "payload",
                label: "Payload",
                value: "",
                placeholder: "Hello LoRaWAN",
                info: {
                  default: "Application payload transmitted by the device",
                },
                error: null,
                required: true,
              }),
              booleanCheckboxField({
                value: false,
                name: "base64Encoded",
                label: "Base64 Encoded",
                error: null,
                text: "Enable",
                info: {
                  default: "Application payload transmitted by the device",
                },
                onChange: (logic: FormLogic) => {
                  if (logic.fieldsState.base64Encoded.value) {
                    logic.setProp(
                      "payload",
                      "placeholder",
                      "SGVsbG8gTG9SQVdBTg==",
                    );
                    logic.setProp("payload", "validations", [
                      {
                        rule: /^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$/,
                        error: "Enter a valid Base64 value",
                      },
                    ]);
                  } else {
                    logic.setProp("payload", "placeholder", "Hello LoRaWAN");
                    logic.setProp("payload", "validations", undefined);
                  }
                },
              }),
              textField({
                name: "antennaRange",
                label: "Antenna Range",
                value: "",
                placeholder: "10000",
                info: {
                  default:
                    "Maximum distance within which the device can communicate, in meters",
                },
                format: (raw: string) => {
                  return raw.replace(/[^0-9]/g, "");
                },
                validations: [
                  {
                    rule: /^[1-9][0-9]*$/,
                    error: "Antenna Range must be greater than 0",
                  },
                ],
                error: null,
                required: true,
              }),
              booleanCheckboxField({
                value: false,
                name: "ADREnabled",
                label: "Adaptive Data Rate",
                error: null,
                text: "Enable",
                info: {
                  default:
                    "Allows the network to optimize the device's data rate and transmission power",
                },
              }),
            ]}
          />
        </div>
        <SensorMap logic={mapLogic} />
      </div>
    </div>
  );
};

export default NewDevicePage;
