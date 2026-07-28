import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import Table from "./Table";
import { TableRecord } from "./Table.types";

const meta = {
  title: "ui-components/widget/Table",
  component: Table,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof Table>;

export default meta;
type Story = StoryObj<typeof meta>;

const testRecords = [
  {
    items: [
      { label: "age", content: "30" },
      { label: "id", content: "0" },
      { label: "name", content: "John" },
      { label: "wrong", content: "Wrong" },
    ],
  },
  {
    items: [
      { label: "id", content: "1" },
      { label: "name", content: "Anna" },
    ],
  },
  {
    items: [
      { label: "id", content: "2" },
      { label: "name", content: "Marco" },
      { label: "age", content: "42" },
    ],
  },
  {
    items: [
      { label: "id", content: "3" },
      { label: "name", content: "Laura" },
      { label: "age", content: "28" },
    ],
  },
  {
    items: [
      { label: "id", content: "4" },
      { label: "name", content: "David" },
      { label: "age", content: "35" },
    ],
  },
  {
    items: [
      { label: "id", content: "5" },
      { label: "name", content: "Sofia" },
      { label: "age", content: "31" },
    ],
  },
  {
    items: [
      { label: "id", content: "6" },
      { label: "name", content: "Luca" },
      { label: "age", content: "22" },
    ],
  },
  {
    items: [
      { label: "id", content: "7" },
      { label: "name", content: "Emma" },
      { label: "age", content: "27" },
    ],
  },
  {
    items: [
      { label: "id", content: "8" },
      { label: "name", content: "Alex" },
      { label: "age", content: "39" },
    ],
  },
  {
    items: [
      { label: "id", content: "9" },
      { label: "name", content: "Giulia" },
      { label: "age", content: "24" },
    ],
  },
  {
    items: [
      { label: "id", content: "10" },
      { label: "name", content: "Tom" },
      { label: "age", content: "45" },
    ],
  },
  {
    items: [
      { label: "id", content: "11" },
      { label: "name", content: "Elena" },
      { label: "age", content: "33" },
    ],
  },
  {
    items: [
      { label: "id", content: "12" },
      { label: "name", content: "Paul" },
      { label: "age", content: "29" },
    ],
  },
  {
    items: [
      { label: "id", content: "13" },
      { label: "name", content: "Sara" },
      { label: "age", content: "36" },
    ],
  },
  {
    items: [
      { label: "id", content: "14" },
      { label: "name", content: "Mike" },
      { label: "age", content: "41" },
    ],
  },
  {
    items: [
      { label: "id", content: "15" },
      { label: "name", content: "Chiara" },
      { label: "age", content: "26" },
    ],
  },
  {
    items: [
      { label: "id", content: "16" },
      { label: "name", content: "Andrea" },
      { label: "age", content: "38" },
    ],
  },
  {
    items: [
      { label: "id", content: "17" },
      { label: "name", content: "Nina" },
      { label: "age", content: "23" },
    ],
  },
  {
    items: [
      { label: "id", content: "18" },
      { label: "name", content: "Robert" },
      { label: "age", content: "50" },
    ],
  },
  {
    items: [
      { label: "id", content: "19" },
      { label: "name", content: "Marta" },
      { label: "age", content: "32" },
    ],
  },
  {
    items: [
      { label: "id", content: "20" },
      { label: "name", content: "Kevin" },
      { label: "age", content: "37" },
    ],
  },
  {
    items: [
      { label: "id", content: "21" },
      { label: "name", content: "Alice" },
      { label: "age", content: "34" },
    ],
  },
  {
    items: [
      { label: "id", content: "22" },
      { label: "name", content: "Fabio" },
      { label: "age", content: "40" },
    ],
  },
  {
    items: [
      { label: "id", content: "23" },
      { label: "name", content: "Maya" },
      { label: "age", content: "21" },
    ],
  },
  {
    items: [
      { label: "id", content: "24" },
      { label: "name", content: "Chris" },
      { label: "age", content: "44" },
    ],
  },
  {
    items: [
      { label: "id", content: "25" },
      { label: "name", content: "Paola" },
      { label: "age", content: "30" },
    ],
  },
  {
    items: [
      { label: "id", content: "26" },
      { label: "name", content: "Leo" },
      { label: "age", content: "27" },
    ],
  },
  {
    items: [
      { label: "id", content: "27" },
      { label: "name", content: "Irene" },
      { label: "age", content: "35" },
    ],
  },
  {
    items: [
      { label: "id", content: "28" },
      { label: "name", content: "Daniel" },
      { label: "age", content: "43" },
    ],
  },
  {
    items: [
      { label: "id", content: "29" },
      { label: "name", content: "Valentina" },
      { label: "age", content: "28" },
    ],
  },
  {
    items: [
      { label: "id", content: "30" },
      { label: "name", content: "Simon" },
      { label: "age", content: "31" },
    ],
  },
  {
    items: [
      { label: "id", content: "31" },
      { label: "name", content: "Clara" },
      { label: "age", content: "26" },
    ],
  },
  {
    items: [
      { label: "id", content: "32" },
      { label: "name", content: "Matteo" },
      { label: "age", content: "39" },
    ],
  },
  {
    items: [
      { label: "id", content: "33" },
      { label: "name", content: "Olivia" },
      { label: "age", content: "29" },
    ],
  },
  {
    items: [
      { label: "id", content: "34" },
      { label: "name", content: "James" },
      { label: "age", content: "46" },
    ],
  },
  {
    items: [
      { label: "id", content: "35" },
      { label: "name", content: "Beatrice" },
      { label: "age", content: "25" },
    ],
  },
  {
    items: [
      { label: "id", content: "36" },
      { label: "name", content: "Marco" },
      { label: "age", content: "48" },
    ],
  },
  {
    items: [
      { label: "id", content: "37" },
      { label: "name", content: "Isabel" },
      { label: "age", content: "37" },
    ],
  },
  {
    items: [
      { label: "id", content: "38" },
      { label: "name", content: "Ryan" },
      { label: "age", content: "33" },
    ],
  },
  {
    items: [
      { label: "id", content: "39" },
      { label: "name", content: "Francesca" },
      { label: "age", content: "42" },
    ],
  },
  {
    items: [
      { label: "id", content: "40" },
      { label: "name", content: "George" },
      { label: "age", content: "36" },
    ],
  },
  {
    items: [
      { label: "id", content: "41" },
      { label: "name", content: "Elisa" },
      { label: "age", content: "24" },
    ],
  },
  {
    items: [
      { label: "id", content: "42" },
      { label: "name", content: "Victor" },
      { label: "age", content: "52" },
    ],
  },
  {
    items: [
      { label: "id", content: "43" },
      { label: "name", content: "Noemi" },
      { label: "age", content: "30" },
    ],
  },
  {
    items: [
      { label: "id", content: "44" },
      { label: "name", content: "Henry" },
      { label: "age", content: "47" },
    ],
  },
  {
    items: [
      { label: "id", content: "45" },
      { label: "name", content: "Alice" },
      { label: "age", content: "28" },
    ],
  },
  {
    items: [
      { label: "id", content: "46" },
      { label: "name", content: "Stefano" },
      { label: "age", content: "34" },
    ],
  },
  {
    items: [
      { label: "id", content: "47" },
      { label: "name", content: "Laura" },
      { label: "age", content: "40" },
    ],
  },
  {
    items: [
      { label: "id", content: "48" },
      { label: "name", content: "Thomas" },
      { label: "age", content: "38" },
    ],
  },
  {
    items: [
      { label: "id", content: "49" },
      { label: "name", content: "Elena" },
      { label: "age", content: "27" },
    ],
  },
];

export const SimpleTable: Story = {
  args: {
    rowLabels: ["id", "name", "age"],
    records: testRecords.slice(0, 5),
  },
};

export const ClickableTable: Story = {
  args: {
    rowLabels: ["id", "name", "age"],
    records: testRecords.slice(0, 5),
    onRowClick: async (row: TableRecord) => {
      await new Promise<void>((resolve) => {
        setTimeout(() => {
          console.table(row);
          resolve();
        }, 2000);
      });
    },
  },
};

export const ScrollableTable: Story = {
  args: {
    rowLabels: ["id", "name", "age"],
    records: testRecords,
  },
  render: (args) => (
    <div style={{ height: "calc(100vh - 4rem)" }}>
      <Table {...args} />
    </div>
  ),
};
