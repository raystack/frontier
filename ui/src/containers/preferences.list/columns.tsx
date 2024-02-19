import {
  V1Beta1Organization,
  V1Beta1Preference,
  V1Beta1PreferenceTrait,
} from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import Skeleton from "react-loading-skeleton";
import { Link } from "react-router-dom";

interface getColumnsOptions {
  traits: V1Beta1PreferenceTrait[];
  preferences: V1Beta1Preference[];
  isLoading?: boolean;
}

export const getColumns: (
  options: getColumnsOptions
) => ColumnDef<V1Beta1PreferenceTrait, any>[] = ({
  traits,
  preferences,
  isLoading,
}) => {
  return [
    {
      header: "Title",
      accessorKey: "title",
      filterVariant: "text",
      cell: isLoading ? () => <Skeleton /> : (info) => info.getValue(),
      footer: (props) => props.column.id,
    },
    {
      header: "Action",
      accessorKey: "name",
      cell: isLoading
        ? () => <Skeleton />
        : (info) => <Link to={`/preferences/${info.getValue()}`}>Edit</Link>,
      footer: (props) => props.column.id,
    },
    {
      header: "Value",
      cell: isLoading
        ? () => <Skeleton />
        : (info) => {
            const name = info.row.original.name;
            const currentPreference =
              name && preferences.find((p) => p.name === name);
            const value =
              (currentPreference && currentPreference.value) ||
              info.row.original.default;
            return value;
          },
      footer: (props) => props.column.id,
    },
  ];
};
