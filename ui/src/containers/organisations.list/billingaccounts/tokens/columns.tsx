import { ApsaraColumnDef, Avatar, Flex, Text } from "@raystack/apsara";
import Skeleton from "react-loading-skeleton";
import { V1Beta1BillingTransaction } from "@raystack/frontier";
import * as R from "ramda";

interface getColumnsOptions {
  isLoading: boolean;
}

const TxnEventSourceMap = {
  "system.starter": "Starter tokens",
  "system.buy": "Recharge",
  "system.awarded": "Complimentary tokens",
  "system.revert": "Refund",
};

export const getColumns: (
  options: getColumnsOptions
) => ApsaraColumnDef<V1Beta1BillingTransaction>[] = ({ isLoading }) => [
  {
    header: "Date",
    accessorKey: "created_at",
    meta: {
      style: {
        paddingLeft: 16,
      },
    },
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          const value = getValue() as Date;
          return (
            <Flex direction="column">
              {new Date(value).toLocaleString("en", {
                month: "long",
                day: "numeric",
                year: "numeric",
              })}
            </Flex>
          );
        },
  },
  {
    header: "Tokens",
    accessorKey: "amount",
    meta: {
      style: {
        paddingLeft: 0,
      },
    },
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          const value = getValue();
          const prefix = row?.original?.type === "credit" ? "+" : "-";
          return (
            <Flex direction="column">
              <Text size={4}>
                {prefix}
                {value}
              </Text>
            </Flex>
          );
        },
  },
  {
    header: "Event",
    accessorKey: "source",
    meta: {
      style: {
        paddingLeft: 0,
      },
    },
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          const value = getValue();
          const eventName = R.pathOr(
            row?.original?.description,
            [value],
            TxnEventSourceMap
          );

          return (
            <Flex direction="column">
              <Text size={4}>{eventName || "-"}</Text>
            </Flex>
          );
        },
  },
  {
    header: "Member",
    accessorKey: "user_id",
    meta: {
      style: {
        minHeight: "48px",
        padding: "12px 0",
      },
    },
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          const userTitle =
            row?.original?.user?.title || row?.original?.user?.email || "-";
          const avatarSrc = row?.original?.user?.avatar;
          return (
            <Flex direction="row" gap={"small"} align={"center"}>
              {avatarSrc ? (
                <Avatar
                  shape={"square"}
                  src={avatarSrc}
                  imageProps={{ width: "24px", height: "24px" }}
                />
              ) : null}
              <Text size={4}>{userTitle}</Text>
            </Flex>
          );
        },
  },
];
