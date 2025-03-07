import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import CreateOrUpdateProduct from "../products.create";
import { V1Beta1Product } from "@raystack/frontier";
import { api } from "~/api";

export default function EditProduct() {
  let { productId } = useParams();
  const [product, setProduct] = useState<V1Beta1Product>();

  useEffect(() => {
    async function fetchProduct() {
      try {
        const res = await api?.frontierServiceGetProduct(productId as string);
        const product = res?.data?.product;
        setProduct(product);
      } catch (error) {
        console.error(error);
      }
    }

    if (productId) fetchProduct();
  }, [productId]);

  return <CreateOrUpdateProduct product={product} />;
}
