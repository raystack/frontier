import {
  FormControl,
  FormField,
  FormFieldProps,
  FormLabel,
  FormMessage,
} from "@radix-ui/react-form";
import { Flex, Text, TextField } from "@raystack/apsara";

import { Control, Controller, UseFormRegister } from "react-hook-form";
import { capitalizeFirstLetter } from "~/utils/helper";

type CustomFieldNameProps = {
  name: string;
  title?: string;
  disabled?: boolean;
  register: UseFormRegister<any>;
  control: Control<any, any>;
};

export const CustomFieldName = ({
  name,
  title,
  register,
  control,
  disabled = false,
  ...props
}: FormFieldProps &
  CustomFieldNameProps &
  React.RefAttributes<HTMLDivElement>) => {
  return (
    <FormField
      name={name}
      defaultValue={props.defaultValue}
      style={{ width: "100%" }}
    >
      <Flex
        gap="medium"
        style={{
          alignItems: "baseline",
          justifyContent: "space-between",
        }}
      >
        <FormLabel>
          <Text>{capitalizeFirstLetter(title || name)}</Text>
        </FormLabel>
        <FormMessage match="valueMissing">Please enter your {name}</FormMessage>
        <FormMessage match="typeMismatch">
          Please provide a valid {title}
        </FormMessage>
      </Flex>
      <FormControl asChild>
        <Controller
          defaultValue={props.defaultValue}
          name={name}
          control={control}
          render={({ field }) => (
            <TextField
              {...field}
              placeholder={`Enter your ${title?.toLowerCase() || name}`}
              disabled={disabled}
            />
          )}
        />
      </FormControl>
    </FormField>
  );
};

const styles = {
  main: { padding: "32px", width: "80%", margin: 0 },
};
