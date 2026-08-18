"use client";

import { SensorMap } from "@/components";
import { DeviceActivation, DeviceClass } from "@lwn-simulator/contracts";
import {
  Form,
  FormField,
  FormValue,
  PageHeader,
  radioField,
  textField,
} from "@lwn-simulator/ui-components";

const NewDevicePage = () => {
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
              error: null,
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
        <SensorMap />
      </div>
    </div>
  );
};

export default NewDevicePage;
