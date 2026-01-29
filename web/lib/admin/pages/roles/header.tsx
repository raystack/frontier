import { DataTable } from "@raystack/apsara";
import { PageHeader } from "../../components/PageHeader";
import styles from "./roles.module.css";

const pageHeader = {
  title: "Roles",
  breadcrumb: [] as { name: string; href?: string }[],
};

export const RolesHeader = () => {
  return (
    <PageHeader
      title={pageHeader.title}
      breadcrumb={pageHeader.breadcrumb}
      className={styles.header}
    >
      <DataTable.Search size="small" placeholder="Search roles..." />
    </PageHeader>
  );
};
