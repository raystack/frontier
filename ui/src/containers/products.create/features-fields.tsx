import { Flex, Label, Select } from "@raystack/apsara";

import RSelect from "react-select";

import { Controller, UseFormReturn } from "react-hook-form";
import { CustomFieldName } from "~/components/CustomField";
import { useFeatures } from "~/hooks/useFeatures";
import { ProductForm, behaviors } from "./contants";

export const FeatureFields = ({
  methods,
}: {
  methods: UseFormReturn<ProductForm>;
}) => {
  const { features } = useFeatures();

  const watchBehavior = methods.watch("behavior", "basic");
  return (
    <Flex direction="column" gap="large">
      <Label size="large">Behavior and features</Label>
      <Flex gap="extra-large" align="end">
        <Controller
          render={({ field }) => (
            <Select onValueChange={(value: any) => field.onChange(value)}>
              <Select.Trigger style={{ height: "26px", width: "100%" }}>
                <Select.Value placeholder="Select Behavior" />
              </Select.Trigger>
              <Select.Content style={{ width: "320px" }}>
                <Select.Group>
                  {behaviors.map((price: { value: string; title: string }) => (
                    <Select.Item value={price.value} key={price.value}>
                      {price.title}
                    </Select.Item>
                  ))}
                </Select.Group>
              </Select.Content>
            </Select>
          )}
          control={methods.control}
          name="behavior"
        />
        {watchBehavior === "per_seat" && (
          <CustomFieldName
            title="Seat limit"
            name={"behavior_config.seat_limit"}
            register={methods.register}
            control={methods.control}
          />
        )}
        {watchBehavior === "credits" && (
          <CustomFieldName
            title="Credits"
            name={"behavior_config.credit_amount"}
            register={methods.register}
            control={methods.control}
          />
        )}
      </Flex>
      <Flex gap="extra-large" align="end">
        <Controller
          render={({ field }) => (
            <RSelect
              isMulti
              placeholder="select multiple features"
              onChange={(data: any) =>
                field.onChange(data.map((d: any) => ({ name: d.value })))
              }
              options={features}
            />
          )}
          control={methods.control}
          name="features"
        />
      </Flex>
      <Flex>
        <CustomFieldName
          title="Add new features (comma separated)"
          name={"newfeatures"}
          register={methods.register}
          control={methods.control}
        />
      </Flex>
    </Flex>
  );
};
