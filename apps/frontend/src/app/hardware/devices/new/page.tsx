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
  FormValue,
  PageHeader,
  radioField,
  selectField,
  textField,
} from "@lwn-simulator/ui-components";

const NewDevicePage = () => {
  // Hooks
  const mapLogic = useSensorMap({ mode: SensorMapMode.Coords });

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
                  "Identifies the application entity handling the device join.",
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
          ]}
        />
        <SensorMap logic={mapLogic} />
      </div>
    </div>
  );
};

export default NewDevicePage;
