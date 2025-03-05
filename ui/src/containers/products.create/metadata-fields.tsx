import { Button, Flex, Label, Text } from "@raystack/apsara/v1";

import { Cross1Icon, PlusIcon } from "@radix-ui/react-icons";
import { UseFormReturn, useFieldArray } from "react-hook-form";
import { CustomFieldName } from "~/components/CustomField";
import { ProductForm } from "./contants";

export const MetadataFields = ({
  methods,
}: {
  methods: UseFormReturn<ProductForm>;
}) => {
  const metadata = useFieldArray({
    control: methods.control,
    name: "metadata",
  });

  return (
    <Flex direction="column" gap="medium">
      <Label size="large">Metadata</Label>
      {metadata.fields.map((item, index) => {
        return (
          <Flex key={item.id} gap="medium" direction="column">
            <Flex gap="medium" align="end">
              <CustomFieldName
                title={`Key ${index + 1}`}
                name={`metadata.${index}.key`}
                register={methods.register}
                control={methods.control}
              />
              <CustomFieldName
                title={`Value ${index + 1}`}
                name={`metadata.${index}.value`}
                register={methods.register}
                control={methods.control}
              />
              {metadata.fields.length !== index + 1 && (
                <Button
                  size="small"
                  style={{ height: "26px" }}
                  onClick={() => metadata.remove(index)}
                  data-test-id="admin-ui-small-cross-btn"
                >
                  <Cross1Icon />
                </Button>
              )}
            </Flex>
            <Flex gap="medium" align="end">
              {metadata.fields.length === index + 1 && (
                <Button
                  size="small"
                  onClick={(event) => {
                    event.preventDefault();
                    metadata.append({
                      key: "",
                      value: "",
                    });
                  }}
                  style={{ width: "100%", height: "26px" }}
                  data-test-id="admin-ui-add-meta"
                >
                  <Flex
                    direction="column"
                    align="center"
                    style={{ paddingRight: "var(--pd-4)" }}
                  >
                    <PlusIcon />
                  </Flex>
                  <Text size={1}>Add new metadata</Text>
                </Button>
              )}
            </Flex>
          </Flex>
        );
      })}
    </Flex>
  );
};
