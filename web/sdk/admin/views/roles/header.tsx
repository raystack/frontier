import type { ReactNode } from "react";
import { DataTable } from "@raystack/apsara-v1";
import { PageHeader } from "../../components/PageHeader";
import styles from "./roles.module.css";

const pageHeader = {
  title: "Roles",
  breadcrumb: [] as { name: string; href?: string }[],
};

export const RolesHeader = ({ icon }: { icon?: ReactNode }) => {
  return (
    <PageHeader
      title={pageHeader.title}
      icon={icon}
      breadcrumb={pageHeader.breadcrumb}
      className={styles.header}
    >
      <DataTable.Search size="small" placeholder="Search roles..." />
    </PageHeader>
  );
};
