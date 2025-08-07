import { zodResolver } from "@hookform/resolvers/zod";

import { Form, FormSubmit } from "@radix-ui/react-form";
import { Button, Flex, Separator, Sheet } from "@raystack/apsara";

import { V1Beta1Feature, V1Beta1Product } from "@raystack/frontier";
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
import { api } from "~/api";

export default function CreateOrUpdateProduct({
  product,
}: {
  product?: V1Beta1Product;
}) {
  let { productId } = useParams();
  const navigate = useNavigate();

  const methods = useForm<ProductForm>({
    resolver: zodResolver(productSchema),
    defaultValues: { ...defaultFormValues },
  });

  const onOpenChange = useCallback(() => {
    navigate("/products");
  }, []);

  useEffect(() => {
    const data = { ...product } as any;
    const metadata = Object.keys(R.pathOr({}, ["metadata"])(data)).map((m) => ({
      key: m,
      value: data.metadata[m],
    }));
    data.metadata = metadata.length ? metadata : [{ key: "", value: "" }];
    data.features = R.pathOr(
      [],
      ["features"],
    )(data).map((f: V1Beta1Feature) => ({ label: f.name, value: f.name }));
    methods.reset(data);
  }, [product]);

  const onSubmit = async (data: any) => {
    try {
      const transformedData = updateResponse(data);
      if (productId) {
        await api?.frontierServiceUpdateProduct(productId, {
          body: transformedData,
        });
      } else {
        await api?.frontierServiceCreateProduct({ body: transformedData });
      }
      toast.success(`${productId ? "product updated" : "product added"}`);
      navigate("/products");
    } catch (error: any) {
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
