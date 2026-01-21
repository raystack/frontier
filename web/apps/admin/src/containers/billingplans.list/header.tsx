import { DataTable } from "@raystack/apsara";
import PageHeader from "~/components/page-header";
import styles from "./plans.module.css";

export const PlanHeader = ({ header }: any) => {
  return (
    <PageHeader
      title={header.title}
      breadcrumb={header?.breadcrumb || []}
      className={styles.header}
    >
      <DataTable.Search size="small" placeholder="Search plans..." />
    </PageHeader>
  );
};
