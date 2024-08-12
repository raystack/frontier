import { zodResolver } from "@hookform/resolvers/zod";
import { Form } from "@radix-ui/react-form";
import {
  Button,
  Flex,
  Label,
  Select,
  Sheet,
  Text,
  TextField,
} from "@raystack/apsara";
import * as z from "zod";

import { useFrontier } from "@raystack/frontier/react";
import { useCallback, useEffect, useState } from "react";
import { Controller, FormProvider, useForm } from "react-hook-form";
import { useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";
import { SheetFooter } from "~/components/sheet/footer";
import { SheetHeader } from "~/components/sheet/header";
import { V1Beta1Product } from "@raystack/frontier";
import Skeleton from "react-loading-skeleton";

const CheckoutSchema = z.object({
  product: z.string(),
  quantity: z.number().min(1),
});
export type CheckoutForm = z.infer<typeof CheckoutSchema>;

export default function AddTokens() {
  const { client } = useFrontier();
  const navigate = useNavigate();
  let { organisationId, billingaccountId } = useParams();
  const [products, setProducts] = useState<V1Beta1Product[]>([]);
  const [isProductsLoading, setIsProductsLoading] = useState(false);
  const [isCheckoutLoading, setIsCheckoutLoading] = useState(false);

  const methods = useForm<CheckoutForm>({
    resolver: zodResolver(CheckoutSchema),
    defaultValues: {
      product: products?.[0]?.id,
      quantity: 1,
    },
  });

  const onOpenChange = useCallback(() => {
    navigate(
      `/organisations/${organisationId}/billingaccounts/${billingaccountId}/tokens`
    );
  }, [billingaccountId, navigate, organisationId]);

  const onSubmit = async (data: CheckoutForm) => {
    try {
      setIsCheckoutLoading(true);
      if (organisationId && billingaccountId) {
        await client?.adminServiceDelegatedCheckout(
          organisationId,
          billingaccountId,
          {
            product_body: {
              ...data,
              quantity: data.quantity.toString(),
            },
          }
        );
        toast.success("tokens added");
        onOpenChange();
        navigate(0);
      }
    } catch (error: any) {
      toast.error("Something went wrong", {
        description: error.message,
      });
    } finally {
      setIsCheckoutLoading(false);
    }
  };

  useEffect(() => {
    async function getProducts() {
      setIsProductsLoading(true);
      try {
        const resp = await client?.frontierServiceListProducts();
        const creditProducts =
          resp?.data?.products?.filter(
            (product) => product.behavior === "credits"
          ) || [];
        setProducts(creditProducts);
      } catch (err: any) {
        console.error(err);
        toast.error("Something went wrong", {
          description: err.message,
        });
      } finally {
        setIsProductsLoading(false);
      }
    }

    getProducts();
  }, [client]);

  return (
    <Sheet open={true}>
      <Sheet.Content
        side="right"
        // @ts-ignore
        style={{
          width: "30vw",
          padding: 0,
          borderRadius: "var(--pd-8)",
          boxShadow: "var(--shadow-sm)",
        }}
        close={false}
      >
        <FormProvider {...methods}>
          <Form onSubmit={methods.handleSubmit(onSubmit)}>
            <SheetHeader
              title="Add tokens"
              onClick={onOpenChange}
            ></SheetHeader>
            <Flex direction="column" gap="medium" style={styles.main}>
              <Label size="large">Product</Label>
              {isProductsLoading ? (
                <Skeleton />
              ) : (
                <Controller
                  name="product"
                  control={methods.control}
                  render={({ field }) => {
                    const { ref, onChange, ...rest } = field;
                    return (
                      <Select
                        {...rest}
                        onValueChange={(value: any) => field.onChange(value)}
                      >
                        <Select.Trigger
                          ref={ref}
                          style={{ height: "26px", width: "100%" }}
                        >
                          <Select.Value placeholder="Select Product" />
                        </Select.Trigger>
                        <Select.Content style={{ width: "320px" }}>
                          <Select.Group>
                            {products.map((p) => (
                              <Select.Item key={p.id} value={p.id || ""}>
                                {p.title || p.name}
                              </Select.Item>
                            ))}
                          </Select.Group>
                        </Select.Content>
                      </Select>
                    );
                  }}
                />
              )}

              <Label size="large">Quantity</Label>
              {isProductsLoading ? (
                <Skeleton />
              ) : (
                <Controller
                  name="quantity"
                  control={methods.control}
                  render={({ field }) => {
                    const { onChange, ...rest } = field;
                    return (
                      <TextField
                        {...rest}
                        type="number"
                        onChange={(e) => onChange(parseInt(e.target.value))}
                        defaultValue={1}
                        min={1}
                      />
                    );
                  }}
                />
              )}
            </Flex>
            <SheetFooter>
              <Button
                type="submit"
                variant="primary"
                disabled={isCheckoutLoading}
              >
                <Text
                  style={{
                    color: "var(--foreground-inverted)",
                  }}
                >
                  {isCheckoutLoading ? "Adding tokens..." : "Add tokens"}
                </Text>
              </Button>
            </SheetFooter>
          </Form>
        </FormProvider>
      </Sheet.Content>
    </Sheet>
  );
}

const styles = {
  main: { padding: "32px", width: "80%" },
};
