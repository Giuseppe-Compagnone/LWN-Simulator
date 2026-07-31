import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import Form from "./Form";
import { textField } from "./components/TextFormField/textField";
import { Card, CardLayout } from "@/components/layout";
import { checkboxField, textAreaField } from "./components";
import { selectField } from "./components/SelectFormField";
import { radioField } from "./components/RadioFormField/RadioField";

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
      selectField({
        name: "hobby",
        label: "Favorite Hobby",
        value: "",
        placeholder: "Your favorite hobby...",
        options: [
          {
            value: "drawing",
            displayed: <>Drawing</>,
          },
          {
            value: "running",
            displayed: (
              <>
                Running
                <span className="material-symbols-outlined">sprint</span>
              </>
            ),
          },
          {
            value: "fishing",
          },
        ],
      }),
      radioField({
        name: "color",
        label: "Favorite Color",
        value: "",
        options: [
          {
            value: "red",
            displayed: <>Red</>,
          },
          {
            value: "green",
            displayed: (
              <>
                Green
                <span
                  className="material-symbols-outlined"
                  style={{ color: "green", fontVariationSettings: "'FILL' 1" }}
                >
                  square
                </span>
              </>
            ),
          },
          {
            value: "blue",
          },
        ],
      }),
      checkboxField({
        name: "skills",
        label: "Your Skills",
        value: [],
        options: [
          {
            value: "html",
            displayed: <>HTML</>,
          },
          {
            value: "css",
            displayed: (
              <>
                CSS
                <span className="material-symbols-outlined">css</span>
              </>
            ),
          },
          {
            value: "js",
          },
        ],
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
