import { DataTable } from "@raystack/apsara";
import { EmptyState, Flex } from "@raystack/apsara/v1";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";

import { V1Beta1Plan } from "@raystack/frontier";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { PlanHeader } from "./header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { api } from "~/api";

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

  const userMapByName = reduceByKey(plans ?? [], "id");

  const tableStyle = plans?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const planList = isPlansLoading
    ? [...new Array(5)].map((_, i) => ({
        name: i.toString(),
        title: "",
      }))
    : plans;

  const columns = getColumns();

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={planList ?? []}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
        isLoading={isPlansLoading}
      >
        <DataTable.Toolbar>
          <PlanHeader header={pageHeader} />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
        <DataTable.DetailContainer>
          <Outlet
            context={{
              user: planId ? userMapByName[planId] : null,
            }}
          />
        </DataTable.DetailContainer>
      </DataTable>
    </Flex>
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
