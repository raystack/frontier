import type { DataTableColumnDef } from "@raystack/apsara/v1";
import type {
  V1Beta1Preference,
  V1Beta1PreferenceTrait,
} from "@raystack/frontier";

import { Link } from "react-router-dom";
import styles from "./preferences.module.css";

interface getColumnsOptions {
  traits: V1Beta1PreferenceTrait[];
  preferences: V1Beta1Preference[];
}

export const getColumns: (
  options: getColumnsOptions,
) => DataTableColumnDef<V1Beta1PreferenceTrait, unknown>[] = ({
  traits,
  preferences,
}) => {
  return [
    {
      header: "Title",
      accessorKey: "title",
      filterVariant: "text",
      cell: (info) => info.getValue(),
      footer: (props) => props.column.id,
    },
    {
      header: "Action",
      accessorKey: "name",
      cell: (info) => <Link to={`/preferences/${info.getValue()}`}>Edit</Link>,
      footer: (props) => props.column.id,
    },
    {
      header: "Value",
      accessorKey: "id",
      classNames: {
        cell: styles.valueColumn,
      },
      cell: (info) => {
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
