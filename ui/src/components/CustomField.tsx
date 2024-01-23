import {
  FormControl,
  FormField,
  FormLabel,
  FormMessage,
} from "@radix-ui/react-form";
import { Flex, Text, TextField } from "@raystack/apsara";

import { Control, Controller, UseFormRegister } from "react-hook-form";
import { capitalizeFirstLetter } from "~/utils/helper";

type CustomFieldNameProps = {
  name: string;
  label?: string;
  register: UseFormRegister<any>;
  defaultValue?: any;
  control: Control<any, any>;
};

export const CustomFieldName = ({
  name,
  label,
  register,
  defaultValue,
  control,
}: CustomFieldNameProps) => {
  return (
    <FormField name={name}>
      <Flex
        gap="medium"
        style={{
          alignItems: "baseline",
          justifyContent: "space-between",
        }}
      >
        <FormLabel>
          <Text>{label || capitalizeFirstLetter(name)}</Text>
        </FormLabel>
        <FormMessage match="valueMissing">Please enter your {name}</FormMessage>
        <FormMessage match="typeMismatch">
          Please provide a valid {name}
        </FormMessage>
      </Flex>
      <FormControl asChild>
        <Controller
          defaultValue={defaultValue}
          name={name}
          control={control}
          rules={{ required: true }}
          render={({ field }) => (
            <TextField {...field} required placeholder={`Enter your ${name}`} />
          )}
        />
      </FormControl>
    </FormField>
  );
};

const styles = {
  main: { padding: "32px", width: "80%", margin: 0 },
};
