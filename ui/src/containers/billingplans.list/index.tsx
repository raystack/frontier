import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";

import { V1Beta1Plan } from "@raystack/frontier";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { PlanHeader } from "./header";

const pageHeader = {
  title: "Plans",
  breadcrumb: [],
};

type ContextType = { plan: V1Beta1Plan | null };
export default function PlanList() {
  const { client } = useFrontier();
  const [plans, setPlans] = useState([]);

  useEffect(() => {
    async function getAllPlans() {
      const {
        // @ts-ignore
        data: { plans },
      } = await client?.frontierServiceListPlans();
      setPlans(plans);
    }
    getAllPlans();
  }, []);

  let { planId } = useParams();

  const userMapByName = reduceByKey(plans ?? [], "id");

  const tableStyle = plans?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={plans ?? []}
        // @ts-ignore
        columns={getColumns(plans)}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <PlanHeader header={pageHeader} />
          <DataTable.FilterChips style={{ paddingTop: "16px" }} />
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
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 plan created</h3>
  </EmptyState>
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
