import { EmptyState, Flex, DataTable } from "@raystack/apsara";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries } from "@raystack/proton/frontier";
import type { Product } from "@raystack/proton/frontier";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { ProductsHeader } from "./header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import styles from "./products.module.css";

type ContextType = { product: Product | null };
export default function ProductList() {
  const {
    data: productsResponse,
    isLoading: isProductsLoading,
    error,
    isError,
  } = useQuery(FrontierServiceQueries.listProducts, {}, {
    staleTime: Infinity,
  });

  const products = productsResponse?.products || [];

  let { productId } = useParams();
  const productMapById = reduceByKey(products ?? [], "id");

  const columns = getColumns();

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <EmptyState
        icon={<ExclamationTriangleIcon />}
        heading="Error Loading Products"
        subHeading={
          error?.message ||
          "Something went wrong while loading products. Please try again."
        }
      />
    );
  }

  return (
    <DataTable
      data={products}
      columns={columns}
      isLoading={isProductsLoading}
      mode="client"
      defaultSort={{ name: "title", order: "asc" }}
    >
      <Flex direction="column" className={styles.tableWrapper}>
        <ProductsHeader />
        <DataTable.Content
          emptyState={noDataChildren}
          classNames={{ root: styles.tableRoot }}
        />
        <Outlet
          context={{
            product: productId ? productMapById[productId] : null,
          }}
        />
      </Flex>
    </DataTable>
  );
}

export function useProduct() {
  return useOutletContext<ContextType>();
}

export const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="0 product created"
    subHeading="Try creating a new product."
  />
);
