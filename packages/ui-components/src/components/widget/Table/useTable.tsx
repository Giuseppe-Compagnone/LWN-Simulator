"use client";

import { useMemo, useState } from "react";
import { TableLogic, TableRecordItemSort, UseTableProps } from "./Table.types";

const useTable = (props: UseTableProps): TableLogic => {
  // States
  const [sortedUp, setSortedUp] = useState<boolean>(true);
  const [currentPage, setCurrentPage] = useState<number>(1);

  // Memo
  const filteredRecords = useMemo(() => {
    const labels = new Set(props.rowLabels.map((label) => label.value));

    return props.records.map((record) => ({
      ...record,
      items: record.items.filter((item) => labels.has(item.label)),
    }));
  }, [props.records, props.rowLabels]);

  const sortedRecords = useMemo(() => {
    if (!props.orderBy) {
      return filteredRecords;
    }

    const label = props.rowLabels.find(
      (label) => label.value === props.orderBy,
    );

    if (!label) {
      return filteredRecords;
    }

    return [...filteredRecords].sort((a, b) => {
      const aValue = a.items.find(
        (item) => item.label === props.orderBy,
      )?.value;

      const bValue = b.items.find(
        (item) => item.label === props.orderBy,
      )?.value;

      if (aValue === undefined) return 1;
      if (bValue === undefined) return -1;

      let result = 0;

      switch (label.sort) {
        case TableRecordItemSort.Alphabetic:
          result = aValue.localeCompare(bValue, "en");
          break;

        case TableRecordItemSort.Numeric:
          result = Number(aValue) - Number(bValue);
          break;
      }

      return sortedUp ? result : -result;
    });
  }, [filteredRecords, props.orderBy, props.rowLabels, sortedUp]);

  const pagesAmount = useMemo(() => {
    if (!props.pageSize) return 1;

    return Math.ceil(props.records.length / props.pageSize);
  }, [props.pageSize, props.records.length]);

  const visiblePages = useMemo(() => {
    return Array.from({ length: 5 }, (_, index) => {
      return currentPage - 2 + index;
    });
  }, [currentPage]);

  // Functions
  const toggleSort = () => {
    setSortedUp((prev) => !prev);
  };

  const nextPage = () => {
    if (currentPage == pagesAmount) return;
    setCurrentPage((prev) => prev + 1);
  };

  const prevPage = () => {
    if (currentPage == 1) return;
    setCurrentPage((prev) => prev - 1);
  };

  return {
    records: props.pageSize
      ? [...sortedRecords].slice(
          (currentPage - 1) * props.pageSize,
          (currentPage - 1) * props.pageSize + props.pageSize,
        )
      : sortedRecords,
    toggleSort,
    sortedUp,
    pagesAmount,
    currentPage,
    nextPage,
    prevPage,
    setCurrentPage,
    visiblePages,
  };
};

export default useTable;
