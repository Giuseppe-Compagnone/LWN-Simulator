import type { Preview } from "@storybook/nextjs-vite";
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-expect-error
import "../src/styles/main.scss";

const preview: Preview = {
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
