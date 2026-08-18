import { JSX } from "react/jsx-runtime";

/**
 * Defines the supported sorting strategies for table columns.
 */
export enum TableRecordItemSort {
  /** Sorts values using numeric comparison. */
  Numeric = 1,

  /** Sorts values using alphabetical comparison. */
  Alphabetic,
}

/**
 * Represents the configuration of a table column label.
 */
export interface TableLabel {
  /**
   * Text displayed as the column header.
   */
  value: string;

  /**
   * Sorting strategy applied to the column.
   *
   * When omitted, the column does not support sorting.
   */
  sort?: TableRecordItemSort;
}

/**
 * Represents a single cell value inside a table row.
 */
export interface TableRecordItem {
  /**
   * Identifier of the cell, matching the related column label.
   */
  label: string;

  /**
   * Raw value associated with the cell.
   */
  value: string;

  /**
   * Custom React element rendered instead of the raw value.
   */
  content?: JSX.Element;
}

/**
 * Represents a single row of data displayed in the table.
 */
export interface TableRecord {
  /**
   * Collection of cells composing the row.
   */
  items: Array<TableRecordItem>;
}

/**
 * Properties for configuring a table component.
 */
export interface TableProps {
  /**
   * Column definitions displayed in the table header.
   */
  rowLabels: Array<TableLabel>;

  /**
   * Data rows rendered inside the table.
   */
  records: Array<TableRecord>;

  /**
   * Column label used as the default sorting key.
   */
  orderBy?: string;

  /**
   * Callback executed when a table row is clicked.
   *
   * Supports both synchronous and asynchronous handlers.
   */
  onRowClick?: (row: TableRecord) => void | Promise<void>;

  /**
   * Number of rows displayed per page.
   *
   * When provided, pagination is enabled and the table displays the configured
   * number of records per page.
   *
   * When omitted, pagination is disabled and all records are displayed on a
   * single page without showing pagination controls.
   */
  pageSize?: number;
  /**
   * Indicates whether the table is currently loading.
   *
   * When enabled, the table displays its loading state instead of the
   * available records.
   *
   * @default false
   */
  isLoading?: boolean;
}

/**
 * Properties accepted by the `useTable` hook.
 *
 * Extends the table configuration properties.
 */
export interface UseTableProps extends TableProps {}

/**
 * State and actions exposed by the table logic hook.
 */
export interface TableLogic {
  /**
   * Current table records after applying sorting and pagination logic.
   */
  records: Array<TableRecord>;

  /**
   * Toggles the current sorting direction.
   */
  toggleSort: () => void;

  /**
   * Indicates whether records are currently sorted in ascending order.
   */
  sortedUp: boolean;

  /**
   * Total number of available pages.
   */
  pagesAmount: number;

  /**
   * Currently selected page index.
   */
  currentPage: number;

  /**
   * Navigates to the next available page.
   */
  nextPage: () => void;

  /**
   * Navigates to the previous available page.
   */
  prevPage: () => void;

  /**
   * Changes the currently selected page.
   *
   * @param page The target page index.
   */
  setCurrentPage: (page: number) => void;

  /**
   * List of page indexes currently visible in the pagination controls.
   */
  visiblePages: Array<number>;
}
