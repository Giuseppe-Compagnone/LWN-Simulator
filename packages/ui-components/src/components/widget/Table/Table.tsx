import { useState } from "react";
import { TableProps } from "./Table.types";
import useTable from "./useTable";
import cn from "classnames";
import { TablePaginator } from "./components";
import { Spinner } from "@/components/common";

const Table = (props: TableProps) => {
  const tableLogic = useTable({ ...props });

  //States
  const [isLoading, setIsLoading] = useState<boolean>(false);

  return (
    <div className="table">
      <table>
        <thead className="table-header">
          <tr>
            {props.rowLabels.map((label, i) => {
              return (
                <th
                  key={i}
                  className={cn("table-label")}
                  onClick={
                    props.orderBy == label.value
                      ? () => {
                          tableLogic.toggleSort();
                        }
                      : undefined
                  }
                  style={{
                    justifyContent:
                      label.value.length > 1 && i == label.value.length - 1
                        ? "flex-end"
                        : "flex-start",
                    cursor:
                      props.orderBy == label.value ? "pointer" : "default",
                  }}
                >
                  {label.value}
                  {props.orderBy == label.value && (
                    <span
                      className="material-symbols-outlined"
                      style={{
                        transform: tableLogic.sortedUp
                          ? "unset"
                          : "rotateX(180deg)",
                      }}
                    >
                      arrow_upward
                    </span>
                  )}
                </th>
              );
            })}
          </tr>
        </thead>
        <tbody
          className="table-rows"
          style={{
            height: props.pageSize ? `${75 * props.pageSize}px` : "unset",
          }}
        >
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
                  const item = row.items.find(
                    (val) => val.label == label.value,
                  );

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
                      {item && (item.content || item.value)}
                    </td>
                  );
                })}
              </tr>
            );
          })}
        </tbody>
        {props.pageSize ? (
          <TablePaginator tableLogic={tableLogic} />
        ) : (
          <div className="table-footer" />
        )}
      </table>
    </div>
  );
};

export default Table;
