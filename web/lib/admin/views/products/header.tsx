import { DataTable } from "@raystack/apsara";
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
}: {
  header?: typeof defaultPageHeader;
  // eslint-disable-next-line no-unused-vars -- callback param name is for type documentation
  onBreadcrumbClick?: (item: { name: string; href?: string }) => void;
}) => {
  return (
    <PageHeader
      title={header.title}
      breadcrumb={header.breadcrumb}
      onBreadcrumbClick={onBreadcrumbClick}
      className={styles.header}
    >
      <DataTable.Search placeholder="Search products..." size="small" />
    </PageHeader>
  );
};
