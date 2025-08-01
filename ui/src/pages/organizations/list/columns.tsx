import {
  Avatar,
  DataTableColumnDef,
  EmptyFilterValue,
  Flex,
  getAvatarColor,
  Text,
} from "@raystack/apsara/v1";
import { V1Beta1Organization, V1Beta1Plan } from "@raystack/frontier";
import type { SearchOrganizationsResponse_OrganizationResult } from "@raystack/proton/frontier";
import {
  isNullTimestamp,
  TimeStamp,
  timestampToDate,
} from "~/utils/connect-timestamp";
import dayjs from "dayjs";
import styles from "./list.module.css";
import { NULL_DATE } from "~/utils/constants";

export const SUBSCRIPTION_STATES = {
  active: "Active",
  past_due: "Past due",
  trialing: "Trialing",
  canceled: "Canceled",
  "": "NA",
} as const;

type SubscriptionState = keyof typeof SUBSCRIPTION_STATES;

interface getColumnsOptions {
  plans: V1Beta1Plan[];
  groupCountMap: Record<string, Record<string, number>>;
}

export const getColumns = ({
  plans,
  groupCountMap,
}: getColumnsOptions): DataTableColumnDef<
  SearchOrganizationsResponse_OrganizationResult,
  unknown
>[] => {
  const planMap = plans.reduce(
    (acc, plan) => {
      const name = plan.name || "";
      acc[name] = `${plan.title} (${plan.interval})`;
      return acc;
    },
    { "": "Standard" } as Record<string, string>,
  );

  return [
    {
      accessorKey: "title",
      header: "Name",
      classNames: {
        cell: styles["name-column"],
        header: styles["name-column"],
      },
      cell: ({ row }) => {
        const avatarColor = getAvatarColor(row?.original?.id || "");
        return (
          <Flex gap={4} align="center">
            <Avatar
              src={row.original.avatar}
              fallback={row.original.title?.[0]}
              color={avatarColor}
            />
            <Text>{row.original.title}</Text>
          </Flex>
        );
      },
      enableColumnFilter: true,
      enableSorting: true,
    },
    {
      accessorKey: "createdBy",
      header: "Creator",
      cell: ({ getValue }) => {
        return getValue();
      },
    },
    {
      accessorKey: "planName",
      header: "Plan",
      cell: ({ getValue }) => {
        return planMap[getValue() as string];
      },
      filterType: "select",
      filterOptions: Object.entries(planMap).map(([value, label]) => ({
        value: value === "" ? EmptyFilterValue : value,
        label,
      })),
      enableColumnFilter: true,
      enableHiding: true,
      enableGrouping: true,
      showGroupCount: true,
      groupCountMap: groupCountMap["planName"] || {},
      groupLabelsMap: planMap,
    },
    {
      accessorKey: "subscriptionCycleEndAt",
      header: "Cycle ends on",
      filterType: "date",
      cell: ({ getValue }) => {
        const value = getValue() as TimeStamp;
        const date = isNullTimestamp(value)
          ? "-"
          : dayjs(timestampToDate(value)).format("YYYY-MM-DD");
        return <Text>{date}</Text>;
      },
      enableColumnFilter: true,
      // enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: "country",
      header: "Country",
      cell: ({ getValue }) => {
        return getValue();
      },
      enableHiding: true,
      classNames: {
        cell: styles["country-column"],
      },
    },
    {
      accessorKey: "paymentMode",
      header: "Payment mode",
      cell: ({ getValue }) => {
        return getValue();
      },
      enableHiding: true,
      defaultHidden: true,
    },
    {
      accessorKey: "subscriptionState",
      header: "Status",
      cell: ({ getValue }) => {
        return SUBSCRIPTION_STATES[getValue() as SubscriptionState];
      },
      filterType: "select",
      filterOptions: Object.entries(SUBSCRIPTION_STATES).map(
        ([value, label]) => ({
          value: value === "" ? EmptyFilterValue : value,
          label,
        }),
      ),
      enableColumnFilter: true,
      enableHiding: true,
      defaultHidden: true,
      enableGrouping: true,
      showGroupCount: true,
      groupCountMap: groupCountMap["subscriptionState"] || {},
      groupLabelsMap: SUBSCRIPTION_STATES,
    },
    {
      accessorKey: "createdAt",
      header: "Created On",
      filterType: "date",
      cell: ({ getValue }) => {
        const value = getValue() as TimeStamp;
        const date = isNullTimestamp(value)
          ? "-"
          : dayjs(timestampToDate(value)).format("YYYY-MM-DD");
        return <Text>{date}</Text>;
      },
      enableHiding: true,
      defaultHidden: true,
      enableSorting: true,
    },
  ];
};
