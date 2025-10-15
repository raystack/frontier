import { zodResolver } from "@hookform/resolvers/zod";

import { Form, FormSubmit } from "@radix-ui/react-form";
import { Button, Flex, Separator, Sheet } from "@raystack/apsara";

import type { Feature, Product } from "@raystack/proton/frontier";
import { useMutation } from "@connectrpc/connect-query";
import { FrontierServiceQueries } from "@raystack/proton/frontier";
import * as R from "ramda";
import { useCallback, useEffect } from "react";
import { FormProvider, useFieldArray, useForm } from "react-hook-form";
import { useNavigate, useParams } from "react-router-dom";
import slugify from "slugify";
import { toast } from "sonner";
import { SheetFooter } from "~/components/sheet/footer";
import { SheetHeader } from "~/components/sheet/header";
import { BaseFields } from "./base-fields";
import { ProductForm, defaultFormValues, productSchema } from "./contants";
import { FeatureFields } from "./features-fields";
import { MetadataFields } from "./metadata-fields";
import { PriceFields } from "./price-fields";
import { updateResponse } from "./transform";

export default function CreateOrUpdateProduct({
  product,
}: {
  product?: Product;
}) {
  let { productId } = useParams();
  const navigate = useNavigate();

  const methods = useForm<ProductForm>({
    resolver: zodResolver(productSchema),
    defaultValues: defaultFormValues,
  });

  const { mutateAsync: createProduct } = useMutation(
    FrontierServiceQueries.createProduct
  );

  const { mutateAsync: updateProduct } = useMutation(
    FrontierServiceQueries.updateProduct
  );

  const onOpenChange = useCallback(() => {
    navigate("/products");
  }, []);

  useEffect(() => {
    if (!product) return;

    const data = { ...product } as any;
    const metadata = Object.keys(data.metadata || {}).map((m) => ({
      key: m,
      value: data.metadata[m],
    }));
    data.metadata = metadata.length ? metadata : [{ key: "", value: "" }];
    data.features = (data.features || []).map((f: Feature) => ({ label: f.name, value: f.name }));

    // Transform prices - convert bigint amount to string for form input
    if (data.prices && data.prices.length > 0) {
      data.prices = data.prices.map((p: any) => ({
        name: p.name,
        interval: p.interval,
        amount: p.amount?.toString() || "",
      }));
    }

    // Transform behaviorConfig - convert bigint fields to string for form input
    if (data.behaviorConfig) {
      data.behaviorConfig = {
        creditAmount: data.behaviorConfig.creditAmount?.toString() || "",
        seatLimit: data.behaviorConfig.seatLimit?.toString() || "",
        minQuantity: data.behaviorConfig.minQuantity?.toString() || "",
        maxQuantity: data.behaviorConfig.maxQuantity?.toString() || "",
      };
    }


    methods.reset(data);
  }, [product]);

  const onSubmit = async (data: any) => {
    try {
      const transformedData = updateResponse(data);
      if (productId) {
        await updateProduct({
          id: productId,
          body: transformedData,
        });
      } else {
        await createProduct({ body: transformedData });
      }
      toast.success(`${productId ? "product updated" : "product added"}`);
      navigate("/products");
    } catch (error: any) {
      console.error("ConnectRPC Error:", error);
      toast.error("Something went wrong", {
        description: error.message,
      });
    }
  };

  const watchTitle = methods.watch("title", "");
  useEffect(() => {
    useFieldArray;
    methods.setValue(
      "name",
      slugify(watchTitle, {
        remove: undefined,
        lower: true,
        strict: false,
        trim: true,
      }),
    );
  }, [watchTitle]);

  return (
    <Sheet open={true}>
      <Sheet.Content
        side="right"
        // @ts-ignore
        style={{
          width: "30vw",
          padding: 0,
          borderRadius: "var(--rs-space-3)",
          boxShadow: "var(--rs-shadow-soft)",
        }}
        close={false}
      >
        <FormProvider {...methods}>
          <Form onSubmit={methods.handleSubmit(onSubmit)}>
            <SheetHeader
              title={productId ? "Update product" : "Add new product"}
              onClick={onOpenChange}
              data-test-id="admin-ui-add-update-product-btn"
            ></SheetHeader>

            <Flex direction="column" gap={9} style={styles.main}>
              <BaseFields methods={methods} />
              <Separator size="full" color="primary" />
              <PriceFields methods={methods} />
              <Separator size="full" color="primary" />

              <FeatureFields methods={methods} />
              <Separator size="full" color="primary" />

              <MetadataFields methods={methods} />
            </Flex>

            <SheetFooter>
              <FormSubmit asChild>
                <Button
                  style={{ height: "inherit" }}
                  data-test-id="admin-ui-add-update-new-product-btn"
                >
                  {productId ? "Update product" : "Add new product"}
                </Button>
              </FormSubmit>
            </SheetFooter>
          </Form>
        </FormProvider>
      </Sheet.Content>
    </Sheet>
  );
}

const styles = {
  main: {
    padding: "32px",
    margin: 0,
    height: "calc(100vh - 120px)",
    overflow: "scroll",
  },
  formfield: {
    marginBottom: "40px",
  },
};
