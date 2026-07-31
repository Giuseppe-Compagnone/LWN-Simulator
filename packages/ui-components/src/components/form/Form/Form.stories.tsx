import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import Form from "./Form";
import { textField } from "./components/TextFormField/textField";
import { Card, CardLayout } from "@/components/layout";
import { textAreaField } from "./components";

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
      textField({
        name: "name",
        label: "Name",
        value: "Name",
        info: "Your Name",
      }),
      textField({
        name: "last",
        label: "Last Name",
        value: "",
        placeholder: "Your last name",
      }),
      textField({
        name: "pass",
        label: "Password",
        value: "",
        placeholder: "Your Password",
        masked: true,
      }),
      textAreaField({
        name: "desc",
        label: "Description",
        value: "",
        placeholder: "About you",
        resize: true,
        charsMax: 20,
      }),
    ],
    onSubmit: (values) => {
      console.table(values);
    },
  },
  render: (args) => (
    <Card layout={CardLayout.Padded}>
      <Form {...args} />
    </Card>
  ),
};
