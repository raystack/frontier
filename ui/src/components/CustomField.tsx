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
  register: UseFormRegister<any>;
  control: Control<any, any>;
};

export const CustomFieldName = ({
  name,
  register,
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
          <Text>{capitalizeFirstLetter(name)}</Text>
        </FormLabel>
        <FormMessage match="valueMissing">Please enter your {name}</FormMessage>
        <FormMessage match="typeMismatch">
          Please provide a valid {name}
        </FormMessage>
      </Flex>
      <FormControl asChild>
        <Controller
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
