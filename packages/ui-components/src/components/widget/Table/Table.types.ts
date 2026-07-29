import { JSX } from "react/jsx-runtime";

export enum TableRecordItemSort {
  Numeric = 1,
  Alphabetic,
}

export interface TableLabel {
  value: string;
  sort?: TableRecordItemSort;
}

export interface TableRecordItem {
  label: string;
  value: string;
  content?: JSX.Element;
}

export interface TableRecord {
  items: Array<TableRecordItem>;
}

export interface TableProps {
  rowLabels: Array<TableLabel>;
  records: Array<TableRecord>;
  orderBy?: string;
  onRowClick?: (row: TableRecord) => void | Promise<void>;
  pageSize?: number;
}

export interface useTableProps extends TableProps {
  records: Array<TableRecord>;
}

export interface TableLogic {
  records: Array<TableRecord>;
  toggleSort: () => void;
  sortedUp: boolean;
}
