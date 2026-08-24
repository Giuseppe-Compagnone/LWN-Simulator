"use client";

import { SensorMap, SensorMapMode, useSensorMap } from "@/components";
import {
  DeviceActivation,
  DeviceClass,
  DeviceRegion,
} from "@lwn-simulator/contracts";
import {
  Form,
  FormField,
  FormLogic,
  FormValue,
  PageHeader,
  radioField,
  selectField,
  SelectOption,
  textField,
} from "@lwn-simulator/ui-components";
import { useEffect, useRef } from "react";

export function estimateRegion(lat: number, lng: number): DeviceRegion | null {
  // Europe
  if (lat >= 35 && lat <= 72 && lng >= -25 && lng <= 45) {
    return DeviceRegion.EU868;
  }

  // Russia
  if (lat >= 41 && lat <= 82 && lng >= 19 && lng <= 180) {
    return DeviceRegion.RU864;
  }

  // South Korea
  if (lat >= 33 && lat <= 39 && lng >= 124 && lng <= 132) {
    return DeviceRegion.KR920;
  }

  // India
  if (lat >= 6 && lat <= 37 && lng >= 68 && lng <= 98) {
    return DeviceRegion.IN865;
  }

  // China
  if (lat >= 18 && lat <= 54 && lng >= 73 && lng <= 135) {
    return DeviceRegion.CN470;
  }

  // Australia
  if (lat >= -44 && lat <= -10 && lng >= 112 && lng <= 154) {
    return DeviceRegion.AU915;
  }

  // USA / Canada / part of Americas
  if (lat >= 24 && lat <= 72 && lng >= -170 && lng <= -50) {
    return DeviceRegion.US915;
  }

  // South America / Japan / Taiwan / SE Asia etc.
  if (lat >= -50 && lat <= 50 && lng >= 95 && lng <= 180) {
    return DeviceRegion.AS923;
  }

  return null;
}

const rx1DataRateOffsetRange: Record<DeviceRegion, [number, number]> = {
  EU868: [0, 7],
  US915: [0, 3],
  CN779: [0, 5],
  EU433: [0, 7],
  AU915: [0, 3],
  CN470: [0, 7],
  AS923: [0, 7],
  KR920: [0, 5],
  IN865: [0, 7],
  RU864: [0, 7],
};

const rx2Frequency: Record<DeviceRegion, number> = {
  EU868: 869525000,
  US915: 923300000,
  CN779: 786000000,
  EU433: 434665000,
  AU915: 923300000,
  CN470: 505300000,
  AS923: 923200000,
  KR920: 923300000,
  IN865: 866550000,
  RU864: 869100000,
};

const NewDevicePage = () => {
  // Hooks
  const formLogicRef = useRef<FormLogic | null>(null);
  const mapLogic = useSensorMap({ mode: SensorMapMode.Coords });

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

      formLogicRef.current.setValue(
        "altitude",
        mapLogic.selectedPos.alt.toString(),
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
        <Form
          onSubmit={(values: Record<string, FormValue>) => {
            console.table(values);
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
              onChange: (logic) => console.table(logic.fieldsState),
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
                default: "Unique identifier assigned to the LoRaWAN end device",
              },
              validations: [
                {
                  rule: /^[0-9A-F]{16}$/,
                  error: "Invalid DevEUI",
                },
              ],
              error: null,
              required: true,
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
                  rule: /^-?(?:90(?:\.0+)?|(?:[0-8]?\d)(?:\.\d+)?)$/,
                  error: "Invalid latitude",
                },
              ],
              required: true,
              onChange: (logic: FormLogic) => {
                if (
                  logic.fieldsState.latitude.value &&
                  (logic.fieldsState.latitude.value as string).match(
                    /^-?(?:90(?:\.0+)?|(?:[0-8]?\d)(?:\.\d+)?)$/,
                  ) &&
                  logic.fieldsState.longitude.value &&
                  (logic.fieldsState.longitude.value as string).match(
                    /^-?(?:180(?:\.0+)?|1[0-7]\d(?:\.\d+)?|[0-9]?\d(?:\.\d+)?)$/,
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
                  rule: /^-?(?:180(?:\.0+)?|1[0-7]\d(?:\.\d+)?|[0-9]?\d(?:\.\d+)?)$/,
                  error: "Invalid longitude",
                },
              ],
              required: true,
              onChange: (logic: FormLogic) => {
                if (
                  logic.fieldsState.latitude.value &&
                  (logic.fieldsState.latitude.value as string).match(
                    /^-?(?:90(?:\.0+)?|(?:[0-8]?\d)(?:\.\d+)?)$/,
                  ) &&
                  logic.fieldsState.longitude.value &&
                  (logic.fieldsState.longitude.value as string).match(
                    /^-?(?:180(?:\.0+)?|1[0-7]\d(?:\.\d+)?|[0-9]?\d(?:\.\d+)?)$/,
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

                  const freq =
                    rx2Frequency[
                      logic.fieldsState.region.value as DeviceRegion
                    ];

                  logic.setProp("rx2CF", "placeholder", freq.toString());
                  logic.setValue("rx2CF", freq.toString());
                } else {
                  logic.setProp("rx1DRO", "options", [] as Array<SelectOption>);
                }
              },
              required: true,
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
                  value: DeviceActivation[key as keyof typeof DeviceActivation],
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
                default: "Session key used to encrypt application payload data",
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
          ]}
        />
        <SensorMap logic={mapLogic} />
      </div>
    </div>
  );
};

export default NewDevicePage;
