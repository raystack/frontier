import type { ReactNode } from "react";
import { DataTable } from "@raystack/apsara-v1";
import { PageHeader } from "../../components/PageHeader";
import styles from "./plans.module.css";

export const PlanHeader = ({ header, icon }: { header: any; icon?: ReactNode }) => {
  return (
    <PageHeader
      title={header.title}
      icon={icon}
      breadcrumb={header?.breadcrumb || []}
      className={styles.header}
    >
      <DataTable.Search size="small" placeholder="Search plans..." />
    </PageHeader>
  );
};
