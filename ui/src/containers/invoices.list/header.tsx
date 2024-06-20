import { DataTable, useTable } from "@raystack/apsara";
import PageHeader from "~/components/page-header";

export const InvoicesHeader = ({ header }: any) => {
  const { filteredColumns } = useTable();
  const isFiltered = filteredColumns.length > 0;

  return (
    <PageHeader title={header.title} breadcrumb={header?.breadcrumb || []}>
      {isFiltered ? <DataTable.ClearFilter /> : <DataTable.FilterOptions />}
      <DataTable.ViewOptions />
      <DataTable.GloabalSearch placeholder="Search invoices..." />
    </PageHeader>
  );
};
