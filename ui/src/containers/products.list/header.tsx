import { DataTable, useTable } from "@raystack/apsara";
import PageHeader from "~/components/page-header";

const defaultPageHeader = {
  title: "Products",
  breadcrumb: [] as {
    href: string;
    name: string;
  }[],
};

export const ProductsHeader = ({ header = defaultPageHeader }) => {
  const { filteredColumns, table } = useTable();
  const isFiltered = filteredColumns.length > 0;

  return (
    <PageHeader title={header.title} breadcrumb={header.breadcrumb}>
      {isFiltered ? <DataTable.ClearFilter /> : <DataTable.FilterOptions />}
      <DataTable.ViewOptions />
      <DataTable.GloabalSearch placeholder="Search products..." />
    </PageHeader>
  );
};
