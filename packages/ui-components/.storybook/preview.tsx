/* eslint-disable @typescript-eslint/ban-ts-comment */
import type { Preview } from "@storybook/nextjs-vite";
// @ts-expect-error
import "../src/styles/main.scss";
// @ts-expect-error
import "./storybook.scss";
import { ArgTypesEnhancer } from "storybook/internal/types";

const enumEnhancer: ArgTypesEnhancer = (context) => {
  const argTypes = context.argTypes;

  Object.entries(argTypes).forEach(([_name, argType]) => {
    const control = argType.control;

    if (
      control &&
      typeof control === "object" &&
      "type" in control &&
      control.type === "object"
    ) {
      const defaultValue = argType.defaultValue;

      if (defaultValue && typeof defaultValue === "object") {
        const values = Object.values(defaultValue);

        if (values.every((value) => typeof value === "string")) {
          argType.control = {
            type: "radio",
          };

          argType.options = values.map((v) => v.replaceAll("-", " ").trim());
        }
      }
    }
  });

  return argTypes;
};

const preview: Preview = {
  argTypesEnhancers: [enumEnhancer],

  globalTypes: {
    theme: {
      name: "Theme",
      description: "Global Theme",
      defaultValue: "light",
      toolbar: {
        icon: "mirror",
        items: ["light", "dark"],
      },
    },
  },

  decorators: [
    (Story, context) => {
      const theme = context.globals.theme;

      return (
        <div className={`sb-wrapper-custom ${theme}`}>
          <Story />
        </div>
      );
    },
  ],
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
};

export default preview;
