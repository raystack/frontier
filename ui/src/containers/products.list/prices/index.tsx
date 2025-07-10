import { DataTable } from "@raystack/apsara/v1";
import { Flex, EmptyState } from "@raystack/apsara/v1";
import { V1Beta1Product } from "@raystack/frontier";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { ProductsHeader } from "../header";
import { getColumns } from "./columns";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { api } from "~/api";

export default function ProductPrices() {
  let { productId } = useParams();
  const [isProductLoading, setIsProductLoading] = useState(false);
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

  const getProduct = useCallback(async () => {
    try {
      setIsProductLoading(true);
      const res = await api?.frontierServiceGetProduct(productId ?? "");
      const product = res?.data?.product;
      setProduct(product);
    } catch (error) {
      console.error(error);
    } finally {
      setIsProductLoading(false);
    }
  }, [productId]);

  useEffect(() => {
    getProduct();
  }, [getProduct]);

  const prices = product?.prices || [];

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={prices}
        columns={getColumns(prices)}
        mode="client"
        isLoading={isProductLoading}
        defaultSort={{ name: "created_at", order: "desc" }}
      >
        <Flex direction="column" width="full">
          <ProductsHeader header={pageHeader} />
          <DataTable.Content emptyState={noDataChildren} />
        </Flex>
      </DataTable>
    </Flex>
  );
}

export const noDataChildren = (
  <EmptyState icon={<ExclamationTriangleIcon />} heading="0 prices found" />
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
