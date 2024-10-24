import { Table } from "@raystack/apsara";
import Skeleton from "react-loading-skeleton";

interface TableLoaderProps {
  row?: number;
  cell?: number;
  cellClassName?: string;
}

export default function TableLoader({
  row = 5,
  cell = 3,
  cellClassName = "",
}: TableLoaderProps) {
  return (
    <>
      {[...new Array(row)].map((_, i) => (
        <Table.Row key={i}>
          {[...new Array(cell)].map((_, j) => (
            <Table.Cell className={cellClassName} key={i + "-" + j}>
              <Skeleton />
            </Table.Cell>
          ))}
        </Table.Row>
      ))}
    </>
  );
}
