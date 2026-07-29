import { useEffect, useState } from "react";
import {
  TableLabel,
  TableLogic,
  TableRecord,
  TableRecordItemSort,
  useTableProps,
} from "./Table.types";

const useTable = (props: useTableProps): TableLogic => {
  // States
  const [records, setRecords] = useState(props.records);
  const [sortedUp, setSortedUp] = useState<boolean>(true);

  // Functions
  const filterRecords = (
    records: Array<TableRecord>,
    rowLabels: Array<TableLabel>,
  ): Array<TableRecord> => {
    return records.map((record) => ({
      ...record,
      items: record.items.filter((item) =>
        rowLabels.map((label) => label.value).includes(item.label),
      ),
    }));
  };

  const toggleSort = () => {
    setSortedUp((prev) => !prev);
  };

  // Effects

  useEffect(() => {
    const updateRecords = async () => {
      setRecords(filterRecords(props.records, props.rowLabels));
    };

    updateRecords();
  }, [props.records, props.rowLabels]);

  useEffect(() => {
    (async () => {
      if (!props.orderBy) return;

      const label = props.rowLabels.find(
        (label) => label.value == props.orderBy,
      );

      if (!label) return;
      switch (label.sort) {
        case TableRecordItemSort.Alphabetic:
          setRecords((prev) => {
            prev.sort((a, b) => {
              const aItem = a.items.find(
                (item) => item.label === props.orderBy,
              );
              const bItem = b.items.find(
                (item) => item.label === props.orderBy,
              );

              if (!aItem && !bItem) return 0;

              if (!aItem) return 1;

              if (!bItem) return -1;

              if (sortedUp) {
                return aItem.value.localeCompare(bItem.value, "en");
              }
              return bItem.value.localeCompare(aItem.value, "en");
            });

            return [...prev];
          });
          break;
        case TableRecordItemSort.Numeric:
          setRecords((prev) => {
            prev.sort((a, b) => {
              const aItem = a.items.find(
                (item) => item.label === props.orderBy,
              );
              const bItem = b.items.find(
                (item) => item.label === props.orderBy,
              );

              if (!aItem && !bItem) return 0;

              if (!aItem) return 1;

              if (!bItem) return -1;

              if (sortedUp) {
                return parseInt(aItem.value) - parseInt(bItem.value);
              }
              return parseInt(aItem.value) + parseInt(bItem.value);
            });

            return [...prev];
          });
          break;
        default:
          break;
      }
    })();
  }, [props.orderBy, props.rowLabels, sortedUp]);

  return {
    records,
    toggleSort,
    sortedUp,
  };
};

export default useTable;
