import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import Form from "./Form";
import { FormFieldType } from "./Form.types";

const meta = {
  title: "ui-components/form/Form",
  component: Form,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof Form>;

export default meta;
type Story = StoryObj<typeof meta>;

export const SimpleForm: Story = {
  args: {
    fields: [
      {
        type: FormFieldType.Text,
        name: "name",
        label: "Name",
        value: "a",
      },
    ],
  },
};
