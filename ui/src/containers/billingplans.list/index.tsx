import { EmptyState, Flex, DataTable, Sheet } from "@raystack/apsara";
import { useEffect, useState } from "react";
import {
  Outlet,
  useNavigate,
  useOutletContext,
  useParams,
} from "react-router-dom";

import type { V1Beta1Plan } from "@raystack/frontier";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { PlanHeader } from "./header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { api } from "~/api";
import styles from "./plans.module.css";
import PageTitle from "~/components/page-title";
import { SheetHeader } from "~/components/sheet/header";

const pageHeader = {
  title: "Plans",
  breadcrumb: [],
};

type ContextType = { plan: V1Beta1Plan | null };
export default function PlanList() {
  const [plans, setPlans] = useState<V1Beta1Plan[]>([]);
  const [isPlansLoading, setIsPlansLoading] = useState(false);

  useEffect(() => {
    async function getAllPlans() {
      setIsPlansLoading(true);
      try {
        const resp = await api?.frontierServiceListPlans();
        const plans = resp?.data?.plans ?? [];
        setPlans(plans);
      } catch (err) {
        console.log(err);
      } finally {
        setIsPlansLoading(false);
      }
    }
    getAllPlans();
  }, []);

  let { planId } = useParams();

  const planMapByName = reduceByKey(plans ?? [], "id");

  const columns = getColumns();

  const navigate = useNavigate();

  function onClose() {
    navigate("/plans");
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
