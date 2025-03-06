import { Button, Flex, Label, Select, Text } from "@raystack/apsara/v1";

import { Cross1Icon, PlusIcon } from "@radix-ui/react-icons";
import { Controller, UseFormReturn, useFieldArray } from "react-hook-form";
import { CustomFieldName } from "~/components/CustomField";
import { ProductForm, intervals } from "./contants";

export const PriceFields = ({
  methods,
}: {
  methods: UseFormReturn<ProductForm>;
}) => {
  const prices = useFieldArray({
    control: methods.control,
    name: "prices",
  });

  return (
    <Flex direction="column" gap="large">
      <Label size="large">Price</Label>
      {prices.fields.map((item, index) => {
        return (
          <Flex key={item.id} gap="extra-large" direction="column">
            <Flex gap="medium" align="end">
              <CustomFieldName
                title="Name (eg. default, monthly)"
                name={`prices.${index}.name`}
                register={methods.register}
                control={methods.control}
              />
              <Controller
                render={({ field }) => (
                  <Select
                    onValueChange={(value: any) => field.onChange(value)}
                    defaultValue={methods.getValues(`prices.${index}.interval`)}
                  >
                    <Select.Trigger style={{ height: "26px", width: "100%" }}>
                      <Select.Value placeholder="select interval" />
                    </Select.Trigger>
                    <Select.Content style={{ width: "320px" }}>
                      <Select.Group>
                        {intervals.map(
                          (price: { value: string; title: string }) => (
                            <Select.Item value={price.value} key={price.value}>
                              {price.title}
                            </Select.Item>
                          )
                        )}
                      </Select.Group>
                    </Select.Content>
                  </Select>
                )}
                control={methods.control}
                name={`prices.${index}.interval`}
              />
            </Flex>
            <Flex gap="medium" align="end">
              <CustomFieldName
                title="Price in dollars"
                name={`prices.${index}.amount`}
                register={methods.register}
                control={methods.control}
              />
              {prices.fields.length === index + 1 ? (
                <Button
                  size="small"
                  style={{ width: "100%", height: "26px" }}
                  data-test-id="admin-ui-price-append-btn"
                  onClick={(event) => {
                    event.preventDefault();
                    prices.append({
                      name: "",
                      interval: "",
                      amount: 0,
                    });
                  }}
                >
                  <Flex
                    direction="column"
                    align="center"
                    style={{ paddingRight: "var(--pd-4)" }}
                  >
                    <PlusIcon />
                  </Flex>
                  <Text size={1}>Add new price</Text>
                </Button>
              ) : (
                <Button
                  size="small"
                  style={{ width: "100%", height: "26px" }}
                  onClick={() => prices.remove(index)}
                  data-test-id="admin-ui-cross-btn"
                >
                  <Flex
                    direction="column"
                    align="center"
                    style={{ paddingRight: "var(--pd-4)" }}
                  >
                    <Cross1Icon />
                  </Flex>
                </Button>
              )}
            </Flex>
          </Flex>
        );
      })}
    </Flex>
  );
};
