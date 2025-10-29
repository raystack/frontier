import {
  Avatar,
  Badge,
  DataTableColumnDef,
  Flex,
  getAvatarColor,
  Text,
} from "@raystack/apsara";
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
import {
  getActionBadgeColor,
  getAuditLogActorName,
  isAuditLogActorSystem,
} from "../util";
import systemIcon from "~/assets/images/system.jpg";
import { OrganizationCell } from "./organization-cell";
import { ComponentPropsWithoutRef } from "react";

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
      cell: ({ getValue }) => {
        const value = getValue() as AuditRecordActor;
        const name = getAuditLogActorName(value);
        const isSystem = isAuditLogActorSystem(value);

        return (
          <Flex gap={4} align="center">
            <Avatar
              size={3}
              fallback={name?.[0]?.toUpperCase()}
              color={getAvatarColor(value?.id ?? "")}
              radius="full"
              src={isSystem ? systemIcon : undefined}
            />
            <Text size="regular">{name}</Text>
          </Flex>
        );
      },
    },
    {
      accessorKey: "orgId",
      header: "Organization",
      classNames: {
        cell: styles["org-column"],
        header: styles["org-column"],
      },
      cell: ({ getValue }) => {
        return <OrganizationCell id={getValue() as string} />;
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
              {value.name}
            </Text>
            <Text size="small" variant="secondary">
              {value.type.toLowerCase()}
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
