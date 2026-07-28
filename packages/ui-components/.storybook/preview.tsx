/* eslint-disable @typescript-eslint/ban-ts-comment */
import type { Preview } from "@storybook/nextjs-vite";
// @ts-expect-error
import "../src/styles/main.scss";
// @ts-expect-error
import "./storybook.scss";

const preview: Preview = {
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
