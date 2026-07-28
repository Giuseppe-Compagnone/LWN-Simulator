import { TableProps } from "./Table.types";
import useTable from "./useTable";

const Table = (props: TableProps) => {
  const tableLogic = useTable({ ...props });

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
        {tableLogic.records.map((row, i) => {
          return (
            <tr key={i} className="table-row">
              {props.rowLabels.map((label, j) => {
                const item = row.items.find((val) => val.label == label);

                if (!item) return <></>;

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
                    {item.content}
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
