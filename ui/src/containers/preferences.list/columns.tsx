import { ApsaraColumnDef } from "@raystack/apsara";
import { V1Beta1Preference, V1Beta1PreferenceTrait } from "@raystack/frontier";

import { Link } from "react-router-dom";

interface getColumnsOptions {
  traits: V1Beta1PreferenceTrait[];
  preferences: V1Beta1Preference[];
}

export const getColumns: (
  options: getColumnsOptions
) => ApsaraColumnDef<V1Beta1PreferenceTrait>[] = ({ traits, preferences }) => {
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
