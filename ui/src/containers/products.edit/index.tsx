import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import CreateOrUpdateProduct from "../products.create";

export default function EditProduct() {
  let { productId } = useParams();
  const { client } = useFrontier();
  const [product, setProduct] = useState();

  useEffect(() => {
    async function fetchProduct() {
      const {
        // @ts-ignore
        data: { product },
      } = await client?.frontierServiceGetProduct(productId as string);
      setProduct(product);
    }

    if (productId) fetchProduct();
  }, [productId]);

  return <CreateOrUpdateProduct product={product} />;
}
