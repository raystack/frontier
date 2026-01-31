import { EmptyState, Flex, DataTable } from "@raystack/apsara";
import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries } from "@raystack/proton/frontier";
import type { Product } from "@raystack/proton/frontier";
import { reduceByKey } from "../../utils/helper";
import { getColumns } from "./columns";
import { ProductsHeader } from "./header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import styles from "./products.module.css";
import { PageTitle } from "../../components/PageTitle";
import ProductDetails from "./details";

export type ProductsPageProps = {
  selectedProductId?: string;
  onSelectProduct?: (productId: string) => void;
  onCloseDetail?: () => void;
  onNavigateToPrices?: (productId: string) => void;
  appName?: string;
};

export default function ProductsPage({
  selectedProductId,
  onSelectProduct,
  onCloseDetail,
  onNavigateToPrices,
  appName,
}: ProductsPageProps = {}) {
  const {
    data: productsResponse,
    isLoading: isProductsLoading,
    error,
    isError,
  } = useQuery(FrontierServiceQueries.listProducts, {}, {
    staleTime: Infinity,
  });

  const products = productsResponse?.products || [];
  const productMapById = reduceByKey(products ?? [], "id");
  const product = selectedProductId ? productMapById[selectedProductId] ?? null : null;

  const columns = getColumns(onNavigateToPrices);

  const handleRowClick = (p: Product) => {
    onSelectProduct?.(p.id ?? "");
  };

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <>
        <PageTitle title="Products" appName={appName} />
        <EmptyState
          icon={<ExclamationTriangleIcon />}
          heading="Error Loading Products"
          subHeading={
            error?.message ||
            "Something went wrong while loading products. Please try again."
          }
        />
      </>
    );
  }

  return (
    <>
      <PageTitle title="Products" appName={appName} />
      <DataTable
        data={products}
        columns={columns}
        isLoading={isProductsLoading}
        mode="client"
        defaultSort={{ name: "title", order: "asc" }}
        onRowClick={handleRowClick}
      >
        <Flex direction="column" className={styles.tableWrapper}>
          <ProductsHeader />
          <DataTable.Content
            emptyState={noDataChildren}
            classNames={{ root: styles.tableRoot }}
          />
          {product && (
            <ProductDetails
              product={product}
              onClose={onCloseDetail ?? (() => {})}
              onNavigateToPrices={onNavigateToPrices ?? (() => {})}
            />
          )}
        </Flex>
      </DataTable>
    </>
  );
}

export const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="0 product created"
    subHeading="Try creating a new product."
  />
);
