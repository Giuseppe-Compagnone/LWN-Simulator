import { useEffect, useState } from "react";
import { TableLogic, TableRecord, useTableProps } from "./Table.types";

const useTable = (props: useTableProps): TableLogic => {
  // States
  const [records, setRecords] = useState(props.records);

  // Functions
  const filterRecords = (
    records: Array<TableRecord>,
    rowLabels: Array<string>,
  ): Array<TableRecord> => {
    return records.map((record) => ({
      ...record,
      items: record.items.filter((item) => rowLabels.includes(item.label)),
    }));
  };

  useEffect(() => {
    const updateRecords = async () => {
      setRecords(filterRecords(props.records, props.rowLabels));
    };

    updateRecords();
  }, [props.records, props.rowLabels]);

  return {
    records,
  };
};

export default useTable;
