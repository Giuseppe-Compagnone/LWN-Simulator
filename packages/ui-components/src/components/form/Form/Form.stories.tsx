import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import Form from "./Form";
import { textField } from "./components/TextFormField/textField";
import { Card, CardLayout } from "@/components/layout";

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
        value: "",
      }),
      textField({
        name: "last",
        label: "Last Name",
        value: "",
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
