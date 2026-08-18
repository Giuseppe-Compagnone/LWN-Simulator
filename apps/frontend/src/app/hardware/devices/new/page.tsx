"use client";

import { SensorMap } from "@/components";
import { DeviceActivation, DeviceClass } from "@lwn-simulator/contracts";
import {
  Form,
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
            }),
            textField({
              name: "dev",
              label: "DevEUI",
              value: "",
              placeholder: "70B3D57ED0001234",
              format: (raw: string) => {
                return raw
                  .replace(/[^0-9A-Fa-f]/g, "")
                  .slice(0, 16)
                  .toUpperCase();
              },
              error: null,
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
                "class-a":
                  "Lowest power consumption, with downlink windows after each uplink",
                "class-b":
                  "Scheduled downlink windows, offering predictable network communication",
                "class-c":
                  "Nearly continuous downlink availability, with highest power consumption",
              },
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
                oota: "Secure activation through network join using unique session credentials",
                abp: "Direct activation using preconfigured device and session credentials",
              },
            }),
          ]}
        />
        <SensorMap />
      </div>
    </div>
  );
};

export default NewDevicePage;
