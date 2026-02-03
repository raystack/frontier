import { ProductPricesView } from "@raystack/frontier/admin";
import { useParams, useNavigate } from "react-router-dom";

export function ProductPricesPageWithRouter() {
  const { productId } = useParams();
  const navigate = useNavigate();

  return (
    <ProductPricesView
      productId={productId ?? ""}
      onBreadcrumbClick={(item) => {
        if (item.href) navigate(item.href);
      }}
    />
  );
}
