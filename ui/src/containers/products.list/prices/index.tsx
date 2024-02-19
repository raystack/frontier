import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { V1Beta1Product } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { ProductsHeader } from "../header";
import { getColumns } from "./columns";

export default function ProductPrices() {
  const { client } = useFrontier();
  let { productId } = useParams();
  const [product, setProduct] = useState<V1Beta1Product>();

  const pageHeader = {
    title: "Products",
    breadcrumb: [
      {
        href: `/products`,
        name: `Product list`,
      },
      {
        href: `/products/${productId}`,
        name: `${productId}`,
      },
      {
        href: "",
        name: `Prices`,
      },
    ],
  };

  useEffect(() => {
    async function getProduct() {
      const {
        // @ts-ignore
        data: { product },
      } = await client?.frontierServiceGetProduct(productId ?? "");
      setProduct(product);
    }
    getProduct();
  }, [productId]);

  const prices = product?.prices || [];
  const tableStyle = prices?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={prices ?? []}
        // @ts-ignore
        columns={getColumns(prices)}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <ProductsHeader header={pageHeader} />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
}

export const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 invoice created</h3>
  </EmptyState>
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
