import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import Logo from "./Logo";

const meta = {
  title: "ui-components/common/Logo",
  component: Logo,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof Logo>;

export default meta;
type Story = StoryObj<typeof meta>;

export const SimpleLogo: Story = {
  args: {},
};
