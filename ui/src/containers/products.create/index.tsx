import { zodResolver } from "@hookform/resolvers/zod";

import { Form, FormSubmit } from "@radix-ui/react-form";
import { Button, Flex, Separator, Sheet, Text } from "@raystack/apsara";

import { V1Beta1Feature, V1Beta1Product } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
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
  product?: V1Beta1Product;
}) {
  let { productId } = useParams();
  const navigate = useNavigate();
  const { client } = useFrontier();

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
      ["features"]
    )(data).map((f: V1Beta1Feature) => ({ label: f.name, value: f.name }));
    methods.reset(data);
  }, [product]);

  const onSubmit = async (data: any) => {
    try {
      const transformedData = updateResponse(data);
      if (productId) {
        await client?.frontierServiceUpdateProduct(productId, {
          body: transformedData,
        });
      } else {
        await client?.frontierServiceCreateProduct({ body: transformedData });
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
      })
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
          borderRadius: "var(--pd-8)",
          boxShadow: "var(--shadow-sm)",
        }}
        close={false}
      >
        <FormProvider {...methods}>
          <Form onSubmit={methods.handleSubmit(onSubmit)}>
            <SheetHeader
              title={productId ? "Update product" : "Add new product"}
              onClick={onOpenChange}
            ></SheetHeader>

            <Flex direction="column" gap="large" style={styles.main}>
              <BaseFields methods={methods} />
              <Separator
                size="full"
                style={{ height: "2px", backgroundColor: "#eee" }}
              />
              <PriceFields methods={methods} />
              <Separator
                size="full"
                style={{ height: "2px", backgroundColor: "#eee" }}
              />

              <FeatureFields methods={methods} />
              <Separator
                size="full"
                style={{ height: "2px", backgroundColor: "#eee" }}
              />
              <MetadataFields methods={methods} />
            </Flex>

            <SheetFooter>
              <FormSubmit asChild>
                <Button variant="primary" style={{ height: "inherit" }}>
                  <Text
                    size={4}
                    style={{ color: "var(--foreground-inverted)" }}
                  >
                    {productId ? "Update product" : "Add new product"}
                  </Text>
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
