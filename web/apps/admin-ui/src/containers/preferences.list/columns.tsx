import type { DataTableColumnDef } from "@raystack/apsara";
import { Preference, PreferenceTrait } from "@raystack/proton/frontier";

import { Link } from "react-router-dom";
import styles from "./preferences.module.css";

interface getColumnsOptions {
  traits: PreferenceTrait[];
  preferences: Preference[];
}

export const getColumns: (
  options: getColumnsOptions,
) => DataTableColumnDef<PreferenceTrait, unknown>[] = ({
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
