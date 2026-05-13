import type { ReactNode } from "react";
import { DataTable } from "@raystack/apsara-v1";
import { PageHeader } from "../../components/PageHeader";
import styles from "./products.module.css";

const defaultPageHeader = {
  title: "Products",
  breadcrumb: [] as {
    href: string;
    name: string;
  }[],
};

export const ProductsHeader = ({
  header = defaultPageHeader,
  onBreadcrumbClick,
  icon,
}: {
  header?: typeof defaultPageHeader;
  // eslint-disable-next-line no-unused-vars -- callback param name is for type documentation
  onBreadcrumbClick?: (item: { name: string; href?: string }) => void;
  icon?: ReactNode;
}) => {
  return (
    <PageHeader
      title={header.title}
      icon={icon}
      breadcrumb={header.breadcrumb}
      onBreadcrumbClick={onBreadcrumbClick}
      className={styles.header}
    >
      <DataTable.Search placeholder="Search products..." size="small" />
    </PageHeader>
  );
};
