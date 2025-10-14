import { DataTable } from "@raystack/apsara";
import { Flex, EmptyState } from "@raystack/apsara";
import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries } from "@raystack/proton/frontier";
import type { Product } from "@raystack/proton/frontier";
import { useParams } from "react-router-dom";
import { ProductsHeader } from "../header";
import { getColumns } from "./columns";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";

export default function ProductPrices() {
  let { productId } = useParams();

  const {
    data: productResponse,
    isLoading: isProductLoading,
    error,
    isError,
  } = useQuery(FrontierServiceQueries.getProduct, { id: productId ?? "" }, {
    staleTime: Infinity,
  });

  const product = productResponse?.product as Product | undefined;

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

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <EmptyState
        icon={<ExclamationTriangleIcon />}
        heading="Error Loading Product"
        subHeading={
          error?.message ||
          "Something went wrong while loading product. Please try again."
        }
      />
    );
  }

  const prices = product?.prices || [];

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={prices}
        columns={getColumns(prices)}
        mode="client"
        isLoading={isProductLoading}
        defaultSort={{ name: "createdAt", order: "desc" }}
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
