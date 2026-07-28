import { JSX } from "react/jsx-runtime";

export interface TableRecordItem {
  label: string;
  content: JSX.Element | string;
}

export interface TableRecord {
  items: Array<TableRecordItem>;
}

export interface TableProps {
  rowLabels: Array<string>;
  records: Array<TableRecord>;
  orderBy?: string;
  onRowClick?: (row: TableRecord) => void | Promise<void>;
}

export interface useTableProps extends TableProps {
  records: Array<TableRecord>;
}

export interface TableLogic {
  records: Array<TableRecord>;
}
