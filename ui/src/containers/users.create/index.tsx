import { zodResolver } from "@hookform/resolvers/zod";
import {
  Button,
  Container,
  Flex,
  Sheet,
  SheetContent,
  Text,
  TextField,
} from "@odpf/apsara";
import {
  Form,
  FormControl,
  FormField,
  FormLabel,
  FormMessage,
  FormSubmit,
} from "@radix-ui/react-form";
import useSWRMutation from "swr/mutation";
import * as z from "zod";

import { useCallback } from "react";
import { FormProvider, useForm, UseFormRegister } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { update } from "~/api";
import { SheetFooter } from "~/components/sheet/footer";
import { SheetHeader } from "~/components/sheet/header";
import { capitalizeFirstLetter } from "~/utils/helper";

const UserSchema = z.object({
  name: z
    .string()
    .trim()
    .min(3, { message: "Must be 3 or more characters long" }),
  email: z.string().email(),
});
export type UserForm = z.infer<typeof UserSchema>;

export default function NewUser() {
  const navigate = useNavigate();
  const { trigger } = useSWRMutation("/admin/v1beta1/users", update, {});

  const methods = useForm<UserForm>({
    resolver: zodResolver(UserSchema),
    defaultValues: {},
  });

  const onOpenChange = useCallback(() => {
    navigate("/users");
  }, []);

  const onSubmit = async (data: any) => {
    await trigger(data);
    navigate("/users");
  };

  return (
    <Sheet open={true}>
      <SheetContent
        side="right"
        css={{
          width: "30vw",
          borderRadius: "$3",
          backgroundColor: "$gray1",
          boxShadow: "0px 0px 6px 1px #E2E2E2",
        }}
        close={false}
      >
        <FormProvider {...methods}>
          <Form onSubmit={methods.handleSubmit(onSubmit)}>
            <SheetHeader
              title="Add new user"
              onClick={onOpenChange}
            ></SheetHeader>
            <Container css={styles.main}>
              <CustomFieldName name="name" register={methods.register} />
              <CustomFieldName name="email" register={methods.register} />
            </Container>
            <SheetFooter>
              <FormSubmit asChild>
                <Button variant="primary" css={{ height: "inherit" }}>
                  <Text
                    css={{
                      fontSize: "14px",
                      color: "white",
                      fontWeight: "normal",
                      lineHeight: "20px",
                      py: "$2",
                    }}
                  >
                    Add user
                  </Text>
                </Button>
              </FormSubmit>
            </SheetFooter>
          </Form>
        </FormProvider>
      </SheetContent>
    </Sheet>
  );
}

type CustomFieldNameProps = {
  name: string;
  register: UseFormRegister<UserForm>;
};

const CustomFieldName = ({ name, register }: CustomFieldNameProps) => {
  return (
    <FormField name={name} style={styles.formfield}>
      <Flex
        css={{
          marginBottom: "$1",
          alignItems: "baseline",
          justifyContent: "space-between",
        }}
      >
        <FormLabel
          style={{ fontSize: "11px", color: "#6F6F6F", lineHeight: "16px" }}
        >
          {capitalizeFirstLetter(name)}
        </FormLabel>
        <FormMessage match="valueMissing">Please enter your {name}</FormMessage>
        <FormMessage match="typeMismatch">
          Please provide a valid {name}
        </FormMessage>
      </Flex>
      <FormControl asChild>
        <TextField
          css={{
            height: "32px",
            color: "$grass12",
            borderRadius: "$3",
            padding: "$2",
          }}
          type="name"
          {...register(name as any)}
          required
          placeholder={`Enter your ${name}`}
        />
      </FormControl>
    </FormField>
  );
};

const styles = {
  main: { padding: "32px", width: "80%", margin: 0 },
  formfield: {
    marginBottom: "40px",
  },
};
