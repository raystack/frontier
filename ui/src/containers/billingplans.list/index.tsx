import { EmptyState, Flex, DataTable, Sheet } from "@raystack/apsara";
import {
  Outlet,
  useNavigate,
  useOutletContext,
  useParams,
} from "react-router-dom";

import type { Plan } from "@raystack/proton/frontier";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { PlanHeader } from "./header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import styles from "./plans.module.css";
import PageTitle from "~/components/page-title";
import { SheetHeader } from "~/components/sheet/header";
import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries } from "@raystack/proton/frontier";

const pageHeader = {
  title: "Plans",
  breadcrumb: [],
};

type ContextType = { plan: Plan | null };
export default function PlanList() {
  const {
    data: plansResponse,
    isLoading: isPlansLoading,
    error,
    isError,
  } = useQuery(FrontierServiceQueries.listPlans, {}, {
    staleTime: Infinity,
  });

  const plans = plansResponse?.plans || [];

  let { planId } = useParams();

  const planMapByName = reduceByKey(plans ?? [], "id");

  const columns = getColumns();

  const navigate = useNavigate();

  function onClose() {
    navigate("/plans");
  }

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <EmptyState
        icon={<ExclamationTriangleIcon />}
        heading="Error Loading Plans"
        subHeading={
          error?.message ||
          "Something went wrong while loading plans. Please try again."
        }
      />
    );
  }

  return (
    <DataTable
      data={plans}
      columns={columns}
      isLoading={isPlansLoading}
      mode="client"
      defaultSort={{ name: "title", order: "asc" }}
    >
      <Flex direction="column">
        <PageTitle title="Plans" />
        <PlanHeader header={pageHeader} />
        <DataTable.Content
          emptyState={noDataChildren}
          classNames={{ root: styles.tableRoot, table: styles.table }}
        />
      </Flex>
      <Sheet open={planId !== undefined}>
        <Sheet.Content className={styles.sheetContent}>
          <SheetHeader title="Plan Details" onClick={onClose} />
          <Flex className={styles.sheetContentBody}>
            <Outlet
              context={{
                plan: planId ? planMapByName[planId] : null,
              }}
            />
          </Flex>
        </Sheet.Content>
      </Sheet>
    </DataTable>
  );
}

export function usePlan() {
  return useOutletContext<ContextType>();
}

export const noDataChildren = (
  <EmptyState icon={<ExclamationTriangleIcon />} heading="0 plan created" />
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
