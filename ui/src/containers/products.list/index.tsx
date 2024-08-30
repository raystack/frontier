import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";

import { V1Beta1Product } from "@raystack/frontier";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { ProductsHeader } from "./header";

type ContextType = { product: V1Beta1Product | null };
export default function ProductList() {
  const { client } = useFrontier();
  const [products, setProducts] = useState<V1Beta1Product[]>([]);
  const [isProductsLoading, setIsProductsLoading] = useState(false);

  useEffect(() => {
    async function getProducts() {
      setIsProductsLoading(true);
      try {
        const {
          // @ts-ignore
          data: { products },
        } = await client?.frontierServiceListProducts() ?? {};
        setProducts(products);
      } catch (err) {
        console.log(err);
      } finally {
        setIsProductsLoading(false);
      }
    }
    getProducts();
  }, [client]);

  let { productId } = useParams();
  const productMapByName = reduceByKey(products ?? [], "id");

  const tableStyle = products?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const productList = isProductsLoading
    ? [...new Array(5)].map((_, i) => ({
        name: i.toString(),
        title: "",
      }))
    : products;

  const columns = getColumns();

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={productList ?? []}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
        isLoading={isProductsLoading}
      >
        <DataTable.Toolbar>
          <ProductsHeader />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
        <DataTable.DetailContainer>
          <Outlet
            context={{
              product: productId ? productMapByName[productId] : null,
            }}
          />
        </DataTable.DetailContainer>
      </DataTable>
    </Flex>
  );
}

export function useProduct() {
  return useOutletContext<ContextType>();
}

export const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 product created</h3>
    <div className="pera">Try creating a new product.</div>
  </EmptyState>
);
