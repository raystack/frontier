import { Flex, Text, Grid, Sheet } from "@raystack/apsara/v1";
import { NavLink, useNavigate } from "react-router-dom";
import { useProduct } from ".";
import styles from "./products.module.css";
import { SheetHeader } from "~/components/sheet/header";

export default function ProductDetails() {
  const { product } = useProduct();
  const navigate = useNavigate();

  function onClose() {
    navigate("/products");
  }

  return (
    <Sheet open>
      <Sheet.Content className={styles.sheetContent}>
        <SheetHeader title="Product Details" onClick={onClose} />
        <Flex
          className={styles.sheetContentBody}
          direction="column"
          gap="large"
        >
          <Text size={4}>{product?.title}</Text>
          <Flex direction="column" gap="large">
            <Grid columns={2} gap="small">
              <Text size={1}>Name</Text>
              <Text size={1}>{product?.title}</Text>
            </Grid>
            <Grid columns={2} gap="small">
              <Text size={1}>Prices</Text>
              <Text size={1}>
                <NavLink to={`/products/${product?.id}/prices`}>
                  Go to prices
                </NavLink>
              </Text>
            </Grid>
          </Flex>
        </Flex>
      </Sheet.Content>
    </Sheet>
  );
}
