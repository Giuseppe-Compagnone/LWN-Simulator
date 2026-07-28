import type { Preview } from "@storybook/nextjs-vite";
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-expect-error
import "../src/styles/main.scss";

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
      document.body.classList.remove("light", "dark");
      document.body.classList.add(context.globals.theme);
      document.body.style.padding = "1rem";

      return <Story />;
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
