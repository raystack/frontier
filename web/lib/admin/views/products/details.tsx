import { Flex, Text, Grid, Sheet } from "@raystack/apsara";
import type { Product } from "@raystack/proton/frontier";
import styles from "./products.module.css";
import { SheetHeader } from "../../components/SheetHeader";

type ProductDetailsProps = {
  product: Product;
  onClose: () => void;
  onNavigateToPrices: (productId: string) => void;
};

export default function ProductDetails({
  product,
  onClose,
  onNavigateToPrices,
}: ProductDetailsProps) {
  return (
    <Sheet open>
      <Sheet.Content className={styles.sheetContent}>
        <SheetHeader title="Product Details" onClick={onClose} />
        <Flex className={styles.sheetContentBody} direction="column" gap={9}>
          <Text size={4}>{product?.title}</Text>
          <Flex direction="column" gap={9}>
            <Grid columns={2} gap="small">
              <Text size={1}>Name</Text>
              <Text size={1}>{product?.title}</Text>
            </Grid>
            <Grid columns={2} gap="small">
              <Text size={1}>Prices</Text>
              <Text size={1}>
                <button
                  type="button"
                  className={styles.linkButton}
                  data-test-id="product-details-prices-link"
                  onClick={() => product?.id && onNavigateToPrices(product.id)}
                >
                  Go to prices
                </button>
              </Text>
            </Grid>
          </Flex>
        </Flex>
      </Sheet.Content>
    </Sheet>
  );
}
