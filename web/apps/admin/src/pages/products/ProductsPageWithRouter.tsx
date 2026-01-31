import { ProductsPage } from "@raystack/frontier/admin";
import { useParams, useNavigate } from "react-router-dom";

export function ProductsPageWithRouter() {
  const { productId } = useParams();
  const navigate = useNavigate();

  return (
    <ProductsPage
      selectedProductId={productId}
      onSelectProduct={(id) => navigate(`/products/${encodeURIComponent(id)}`)}
      onCloseDetail={() => navigate("/products")}
      onNavigateToPrices={(id) => navigate(`/products/${encodeURIComponent(id)}/prices`)}
    />
  );
}
