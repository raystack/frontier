import {
  FormControl,
  FormField,
  FormFieldProps,
  FormLabel,
  FormMessage,
} from "@radix-ui/react-form";
import { Flex, Select, Switch, Text, TextField } from "@raystack/apsara";
import React, { CSSProperties } from "react";

import { Control, Controller, UseFormRegister } from "react-hook-form";
import { capitalizeFirstLetter } from "~/utils/helper";
import { MultiSelect } from "./multiselect";
import Skeleton from "react-loading-skeleton";

type CustomFieldNameProps = {
  name: string;
  isLoading?: boolean;
  title?: string;
  disabled?: boolean;
  register: UseFormRegister<any>;
  control: Control<any, any>;
  variant?: "textarea" | "input" | "select" | "multiselect" | "switch";
  style?: CSSProperties;
  options?: Array<{ label: string; value: any }>;
  placeholder?: string;
};

export const CustomFieldName = ({
  name,
  title,
  control,
  disabled = false,
  variant = "input",
  style = {},
  placeholder,
  options = [],
  isLoading = false,
  ...props
}: FormFieldProps &
  CustomFieldNameProps &
  React.RefAttributes<HTMLDivElement>) => {
  const inputTitle = capitalizeFirstLetter(title || name);
  return (
    <FormField
      name={name}
      defaultValue={props.defaultValue}
      style={{ width: "100%" }}
      asChild
    >
      <Flex direction="column" gap="small">
        <Flex
          gap="medium"
          style={{
            alignItems: "baseline",
            justifyContent: "space-between",
          }}
        >
          <FormLabel>
            <Text>{inputTitle}</Text>
          </FormLabel>
          <FormMessage match="valueMissing">
            Please enter your {name}
          </FormMessage>
          <FormMessage match="typeMismatch">
            Please provide a valid {title}
          </FormMessage>
        </Flex>
        <FormControl asChild>
          {isLoading ? (
            <Skeleton />
          ) : (
            <Controller
              defaultValue={props.defaultValue}
              name={name}
              control={control}
              render={({ field }) => {
                switch (variant) {
                  case "textarea": {
                    return (
                      <textarea
                        {...field}
                        defaultValue={props?.defaultValue}
                        placeholder={
                          placeholder ||
                          `Enter your ${title?.toLowerCase() || name}`
                        }
                        style={style}
                      />
                    );
                  }
                  case "select": {
                    const { ref, ...rest } = field;
                    return (
                      <Select
                        {...rest}
                        onValueChange={(value: any) => field.onChange(value)}
                      >
                        <Select.Trigger
                          ref={ref}
                          style={{ height: "26px", width: "100%" }}
                        >
                          <Select.Value placeholder={`Select ${inputTitle}`} />
                        </Select.Trigger>
                        <Select.Content style={{ width: "320px" }}>
                          <Select.Group>
                            {options.map((opt) => (
                              <Select.Item key={opt.value} value={opt.value}>
                                {opt.label}
                              </Select.Item>
                            ))}
                          </Select.Group>
                        </Select.Content>
                      </Select>
                    );
                  }
                  case "multiselect": {
                    const { onChange, value, ...rest } = field;
                    return (
                      <MultiSelect<string>
                        {...rest}
                        options={options}
                        onSelect={onChange}
                        selected={value || props?.defaultValue}
                      />
                    );
                  }
                  case "switch": {
                    const { onChange, value, ...rest } = field;
                    return (
                      <Switch
                        {...rest}
                        defaultChecked={props?.defaultChecked}
                        checked={value}
                        onCheckedChange={onChange}
                      />
                    );
                  }

                  default: {
                    return (
                      <TextField
                        {...field}
                        placeholder={
                          placeholder ||
                          `Enter your ${title?.toLowerCase() || name}`
                        }
                        disabled={disabled}
                      />
                    );
                  }
                }
              }}
            />
          )}
        </FormControl>
      </Flex>
    </FormField>
  );
};
