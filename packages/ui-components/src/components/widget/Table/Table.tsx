import { useState } from "react";
import { TableProps } from "./Table.types";
import useTable from "./useTable";
import { Spinner } from "@/components/common";
import cn from "classnames";

const Table = (props: TableProps) => {
  const tableLogic = useTable({ ...props });

  //States
  const [isLoading, setIsLoading] = useState<boolean>(false);

  return (
    <table className="table">
      <thead className="table-header">
        <tr>
          {props.rowLabels.map((label, i) => {
            return (
              <th
                key={i}
                className="table-label"
                style={{
                  textAlign:
                    label.length > 1 && i == label.length - 1
                      ? "right"
                      : "left",
                }}
              >
                {label}
              </th>
            );
          })}
        </tr>
      </thead>
      <tbody className="table-rows">
        <div className={cn("loading-wrapper", isLoading && "visible")}>
          {isLoading && <Spinner />}
        </div>
        {tableLogic.records.map((row, i) => {
          return (
            <tr
              key={i}
              className={cn("table-row", props.onRowClick && "clickable")}
              onClick={async () => {
                setIsLoading(true);
                await props.onRowClick?.(row);
                setIsLoading(false);
              }}
            >
              {props.rowLabels.map((label, j) => {
                const item = row.items.find((val) => val.label == label);

                return (
                  <td
                    key={j}
                    className="table-row-item"
                    style={{
                      justifyContent:
                        props.rowLabels.length > 1 &&
                        j == props.rowLabels.length - 1
                          ? "flex-end"
                          : "flex-start",
                    }}
                  >
                    {item && item.content}
                  </td>
                );
              })}
            </tr>
          );
        })}
      </tbody>
    </table>
  );
};

export default Table;
