import { V1Beta1Plan, V1Beta1Subscription } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { Text } from "@raystack/apsara";
const columnHelper = createColumnHelper<V1Beta1Subscription>();

interface getColumnsOptions {
  subscriptions: V1Beta1Subscription[];
  plans: V1Beta1Plan[];
}
export const getColumns: (
  opts: getColumnsOptions
) => ColumnDef<V1Beta1Subscription, any>[] = ({ subscriptions, plans }) => {
  const plansMap = plans.reduce((acc, plan) => {
    const planId = plan.id || "";
    acc[planId] = plan;
    return acc;
  }, {} as Record<string, V1Beta1Plan>);
  return [
    {
      header: "Provider Id",
      accessorKey: "provider_id",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Plan",
      accessorKey: "plan_id",
      cell: (info) => {
        const planId = info.getValue();
        const planName = `${plansMap[planId]?.title} (${plansMap[planId]?.interval})`;
        return <Text>{planName}</Text>;
      },
      filterVariant: "text",
    },
    {
      header: "Period start date",
      accessorKey: "current_period_start_at",
      meta: {
        headerFilter: false,
      },
      cell: (info) =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),

      footer: (props) => props.column.id,
    },
    {
      header: "Period end date",
      accessorKey: "current_period_end_at",
      meta: {
        headerFilter: false,
      },
      cell: (info) =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),

      footer: (props) => props.column.id,
    },
  ];
};
