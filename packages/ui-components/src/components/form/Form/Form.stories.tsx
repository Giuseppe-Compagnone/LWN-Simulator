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
        error: null,
      }),
      textField({
        name: "last",
        label: "Last Name",
        value: "",
        placeholder: "Your last name",
        error: null,
      }),
      textField({
        name: "pass",
        label: "Password",
        value: "",
        placeholder: "Your Password",
        error: null,
        masked: true,
      }),
      textAreaField({
        name: "desc",
        label: "Description",
        value: "",
        placeholder: "About you",
        error: null,
        resize: true,
        charsMax: 20,
      }),
      selectField({
        name: "hobby",
        label: "Favorite Hobby",
        value: "",
        placeholder: "Your favorite hobby...",
        error: null,
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
        error: null,
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
        error: null,
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

export const ErrorForm: Story = {
  args: {
    fields: [
      textField({
        name: "name",
        label: "Name",
        value: "Name",
        info: "Your Name",
        error: null,
        required: true,
      }),
      textField({
        name: "last",
        label: "Last Name",
        value: "",
        placeholder: "Your last name",
        error: null,
        format: (raw) => raw.toUpperCase(),
      }),
      textField({
        name: "pass",
        label: "Password",
        value: "",
        placeholder: "Your Password",
        error: null,
        masked: true,
      }),
      textAreaField({
        name: "desc",
        label: "Description",
        value: "",
        placeholder: "About you",
        error: null,
        resize: true,
        charsMax: 20,
        required: true,
        format: (raw) => raw.replaceAll(/[0-9]/g, ""),
      }),
      selectField({
        name: "hobby",
        label: "Favorite Hobby",
        value: "",
        placeholder: "Your favorite hobby...",
        error: null,
        required: true,

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
        error: null,
        required: true,

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
        error: null,
        required: true,

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
