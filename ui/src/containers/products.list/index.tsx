import { EmptyState, Flex, DataTable, Sheet } from "@raystack/apsara/v1";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";

import type { V1Beta1Product } from "@raystack/frontier";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { ProductsHeader } from "./header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { api } from "~/api";
import styles from "./products.module.css";

type ContextType = { product: V1Beta1Product | null };
export default function ProductList() {
  const [products, setProducts] = useState<V1Beta1Product[]>([]);
  const [isProductsLoading, setIsProductsLoading] = useState(false);

  useEffect(() => {
    async function getProducts() {
      setIsProductsLoading(true);
      try {
        const res = await api?.frontierServiceListProducts();
        const products = res?.data?.products ?? [];
        setProducts(products);
      } catch (err) {
        console.error(err);
      } finally {
        setIsProductsLoading(false);
      }
    }
    getProducts();
  }, []);

  let { productId } = useParams();
  const productMapByName = reduceByKey(products ?? [], "id");

  const columns = getColumns();

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
            product: productId ? productMapByName[productId] : null,
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
