import type { ReactNode } from "react";
import { Flex, EmptyState, DataTable } from "@raystack/apsara";
import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries, type Product } from "@raystack/proton/frontier";
import { PageHeader } from "../../../components/PageHeader";
import { getColumns } from "./columns";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { PageTitle } from "../../../components/PageTitle";
import styles from "../products.module.css";

export type ProductPricesViewProps = {
  /** The product ID whose prices are displayed. */
  productId: string;
  /** App name displayed in the page title. */
  appName?: string;
  /** Called when a breadcrumb item is clicked. Use to navigate back (e.g. to products list). */
  onBreadcrumbClick?: (item: { name: string; href?: string }) => void;
};

export default function ProductPricesView({
  productId,
  appName,
  onBreadcrumbClick,
}: ProductPricesViewProps) {
  const {
    data: productResponse,
    isLoading: isProductLoading,
    error,
    isError,
  } = useQuery(
    FrontierServiceQueries.getProduct,
    { id: productId ?? "" },
    { staleTime: Infinity }
  );

  const product = productResponse?.product as Product | undefined;

  const pageHeader = {
    title: "Products",
    breadcrumb: [
      { href: `/products`, name: "Product list" },
      { href: `/products/${productId}`, name: `${productId}` },
      { href: "", name: "Prices" },
    ],
  };

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <>
        <PageTitle title="Product Prices" appName={appName} />
        <EmptyState
          icon={<ExclamationTriangleIcon />}
          heading="Error Loading Product"
          subHeading={
            error?.message ||
            "Something went wrong while loading product. Please try again."
          }
        />
      </>
    );
  }

  const prices = product?.prices || [];

  return (
    <>
      <PageTitle title="Product Prices" appName={appName} />
      <Flex direction="row" style={{ height: "100%", width: "100%" }}>
        <DataTable
          data={prices}
          columns={getColumns(prices)}
          mode="client"
          isLoading={isProductLoading}
          defaultSort={{ name: "createdAt", order: "desc" }}
        >
          <Flex direction="column" style={{ width: "100%" }}>
            <PageHeader
              title={pageHeader.title}
              breadcrumb={pageHeader.breadcrumb}
              onBreadcrumbClick={onBreadcrumbClick}
              className={styles.header}
            >
              <DataTable.Search placeholder="Search products..." size="small" />
            </PageHeader>
            <DataTable.Content emptyState={noDataChildren} />
          </Flex>
        </DataTable>
      </Flex>
    </>
  );
}

export const noDataChildren = (
  <EmptyState icon={<ExclamationTriangleIcon />} heading="0 prices found" />
);

export const TableDetailContainer = ({ children }: { children: ReactNode }) => (
  <div>{children}</div>
);
