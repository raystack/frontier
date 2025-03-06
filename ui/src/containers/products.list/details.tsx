import { Grid } from "@raystack/apsara";
import { Flex, Text } from "@raystack/apsara/v1";
import { NavLink } from "react-router-dom";
import { useProduct } from ".";

export default function ProductDetails() {
  const { product } = useProduct();

  return (
    <Flex
      direction="column"
      gap="large"
      style={{
        width: "320px",
        height: "calc(100vh - 60px)",
        borderLeft: "1px solid var(--border-base)",
        padding: "var(--pd-16)",
      }}
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
  );
}
