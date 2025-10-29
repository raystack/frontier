import { Badge, DataTableColumnDef, Flex, Text } from "@raystack/apsara";
import dayjs from "dayjs";
import styles from "./list.module.css";
import {
  AuditRecord,
  AuditRecordActor,
  AuditRecordResource,
} from "@raystack/proton/frontier";
import {
  isNullTimestamp,
  TimeStamp,
  timestampToDate,
} from "~/utils/connect-timestamp";
import { ACTOR_TYPES, getActionBadgeColor } from "../util";
import { ComponentPropsWithoutRef } from "react";
import ActorCell from "./actor-cell";

interface getColumnsOptions {
  groupCountMap: Record<string, Record<string, number>>;
}

export const getColumns = ({
  groupCountMap,
}: getColumnsOptions): DataTableColumnDef<AuditRecord, unknown>[] => {
  return [
    {
      accessorKey: "actor",
      header: "Actor",
      classNames: {
        cell: styles["name-column"],
        header: styles["name-column"],
      },
      enableColumnFilter: true,
      cell: ({ getValue }) => (
        <ActorCell value={getValue() as AuditRecordActor} />
      ),
    },
    {
      accessorKey: "actorType",
      header: "Actor Type",
      enableColumnFilter: true,
      defaultHidden: true,
      cell: () => null,
      filterType: "multiselect",
      filterOptions: [
        { label: "User", value: ACTOR_TYPES.USER },
        { label: "Service User", value: ACTOR_TYPES.SERVICE_USER },
        { label: "System", value: ACTOR_TYPES.SYSTEM },
      ],
    },
    {
      accessorKey: "orgName",
      header: "Organization",
      classNames: {
        cell: styles["org-column"],
        header: styles["org-column"],
      },
      cell: ({ getValue }) => {
        return (
          <Text size="regular" className={styles.capitalize}>
            {(getValue() as string) || "-"}
          </Text>
        );
      },
      enableColumnFilter: true,
    },
    {
      accessorKey: "event",
      header: "Action",
      cell: ({ getValue }) => {
        const value = getValue() as string;
        const color = getActionBadgeColor(value) as ComponentPropsWithoutRef<
          typeof Badge
        >["variant"];
        return <Badge variant={color}>{value}</Badge>;
      },
      enableColumnFilter: true,
      enableSorting: true,
    },
    {
      accessorKey: "resource",
      header: "Resource",
      enableColumnFilter: true,
      cell: ({ getValue }) => {
        const value = getValue() as AuditRecordResource;
        return (
          <Flex gap={1} direction="column">
            <Text size="small" weight="medium">
              {value.name || "-"}
            </Text>
            <Text size="small" variant="secondary">
              {value.type.toLowerCase() ?? "-"}
            </Text>
          </Flex>
        );
      },
    },
    {
      accessorKey: "resourceType",
      header: "Resource Type",
      enableColumnFilter: true,
      defaultHidden: true,
      cell: () => null,
    },
    {
      accessorKey: "occurredAt",
      header: "Timestamp",
      filterType: "date",
      cell: ({ getValue }) => {
        const value = getValue() as TimeStamp;
        if (isNullTimestamp(value)) {
          return <Text>-</Text>;
        }
        const date = dayjs(timestampToDate(value));
        return (
          <Flex gap={1} direction="column">
            <Text size="small" weight="medium">
              {date.format("DD MMM YYYY")}
            </Text>
            <Text size="small" variant="secondary">
              {date.format("hh:mm A")}
            </Text>
          </Flex>
        );
      },
      enableHiding: true,
      enableSorting: true,
    },
  ];
};
