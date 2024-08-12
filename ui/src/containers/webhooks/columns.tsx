import { ApsaraColumnDef } from "@raystack/apsara";
import { V1Beta1Webhook } from "@raystack/frontier";

export const getColumns: () => ApsaraColumnDef<V1Beta1Webhook>[] = () => {
  return [
    {
      header: "Description",
      accessorKey: "description",
      filterVariant: "text",
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "State",
      accessorKey: "state",
      filterVariant: "text",
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "URL",
      accessorKey: "url",
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "Created at",
      accessorKey: "created_at",
      cell: (info) =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),
    },
    {
      header: "Action",
      accessorKey: "id",
      cell: (info) => null,
    },
  ];
};
