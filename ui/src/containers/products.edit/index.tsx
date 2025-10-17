import { useParams } from "react-router-dom";
import CreateOrUpdateProduct from "../products.create";
import type { Product } from "@raystack/proton/frontier";
import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries } from "@raystack/proton/frontier";
import { EmptyState } from "@raystack/apsara";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";

export default function EditProduct() {
  let { productId } = useParams();

  const {
    data: productResponse,
    isLoading,
    error,
    isError,
  } = useQuery(
    FrontierServiceQueries.getProduct,
    { id: productId ?? "" },
    {
      staleTime: Infinity,
      enabled: !!productId,
    }
  );

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

  const product = productResponse?.product as Product | undefined;

  return <CreateOrUpdateProduct product={product} isLoading={isLoading} />;
}
