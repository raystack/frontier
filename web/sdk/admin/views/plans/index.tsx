import { EmptyState, Flex, DataTable, Sheet } from "@raystack/apsara";
import type { Plan } from "@raystack/proton/frontier";
import { reduceByKey } from "../../utils/helper";
import { getColumns } from "./columns";
import { PlanHeader } from "./header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import styles from "./plans.module.css";
import { PageTitle } from "../../components/PageTitle";
import { SheetHeader } from "../../components/SheetHeader";
import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries } from "@raystack/proton/frontier";
import PlanDetails from "./details";

const pageHeader = {
  title: "Plans",
  breadcrumb: [],
};

export type PlansViewProps = {
  selectedPlanId?: string;
  onCloseDetail?: () => void;
  appName?: string;
};

export default function PlansView({
  selectedPlanId,
  onCloseDetail,
  appName,
}: PlansViewProps = {}) {
  const {
    data: plansResponse,
    isLoading: isPlansLoading,
    error,
    isError,
  } = useQuery(FrontierServiceQueries.listPlans, {}, {
    staleTime: Infinity,
  });

  const plans = plansResponse?.plans || [];
  const planMapById = reduceByKey(plans ?? [], "id");
  const columns = getColumns();

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
        <PageTitle title="Plans" appName={appName} />
        <PlanHeader header={pageHeader} />
        <DataTable.Content
          emptyState={noDataChildren}
          classNames={{ root: styles.tableRoot, table: styles.table }}
        />
      </Flex>
      <Sheet open={selectedPlanId !== undefined}>
        <Sheet.Content className={styles.sheetContent}>
          <SheetHeader title="Plan Details" onClick={onCloseDetail ?? (() => {})} />
          <Flex className={styles.sheetContentBody}>
            <PlanDetails
              plan={selectedPlanId ? planMapById[selectedPlanId] ?? null : null}
            />
          </Flex>
        </Sheet.Content>
      </Sheet>
    </DataTable>
  );
}

export const noDataChildren = (
  <EmptyState icon={<ExclamationTriangleIcon />} heading="0 plan created" />
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
