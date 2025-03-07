import { DataTable } from "@raystack/apsara";
import { Flex, EmptyState } from "@raystack/apsara/v1";
import { V1Beta1Product } from "@raystack/frontier";
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { ProductsHeader } from "../header";
import { getColumns } from "./columns";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { api } from "~/api";

export default function ProductPrices() {
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

  async function getProduct() {
    try {
      const res = await api?.frontierServiceGetProduct(productId ?? "");
      const product = res?.data?.product;
      setProduct(product);
    } catch (error) {
      console.error(error);
    }
  }

  useEffect(() => {
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
  <EmptyState icon={<ExclamationTriangleIcon />} heading="0 invoice created" />
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
