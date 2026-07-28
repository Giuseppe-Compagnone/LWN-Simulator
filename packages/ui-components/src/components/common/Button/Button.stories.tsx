import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import Button from "./Button";
import { ButtonFont, ButtonType } from "./Button.types";

const meta = {
  title: "ui-components/common/Button",
  component: Button,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

export const SimpleButton: Story = {
  args: {
    value: "Button Text",
    onClick: () => {
      console.log("Click");
    },
  },
};

export const SecondaryButton: Story = {
  args: {
    value: "Button Text",
    type: ButtonType.Secondary,
    onClick: (e) => {
      console.log("TARGET", e?.target);
    },
  },
};

export const OutlinedButton: Story = {
  args: {
    value: "Button Text",
    type: ButtonType.Outlined,
  },
};

export const BlockButton: Story = {
  args: {
    value: "Button Text",
    type: ButtonType.Outlined,
    block: true,
    font: ButtonFont.Secondary,
  },
};

export const LoadingButton: Story = {
  args: {
    value: "Button Text",
    type: ButtonType.Primary,
    font: ButtonFont.Secondary,
    onClick: async () => {
      await new Promise<void>((resolve) => {
        setTimeout(() => {
          resolve();
        }, 2000);
      });
    },
  },
};
