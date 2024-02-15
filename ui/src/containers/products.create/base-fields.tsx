import { Flex } from "@raystack/apsara";
import { UseFormReturn } from "react-hook-form";

import { CustomFieldName } from "~/components/CustomField";
import { ProductForm } from "./contants";

export const BaseFields = ({
  methods,
}: {
  methods: UseFormReturn<ProductForm>;
}) => {
  return (
    <Flex direction="column" gap="large">
      <Flex gap="extra-large">
        <CustomFieldName
          name="title"
          register={methods.register}
          control={methods.control}
        />
        <CustomFieldName
          name="name"
          register={methods.register}
          control={methods.control}
          disabled
        />
      </Flex>
      <CustomFieldName
        name="description"
        register={methods.register}
        control={methods.control}
      />
    </Flex>
  );
};
