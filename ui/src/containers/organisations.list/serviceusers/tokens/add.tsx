import { Form, FormSubmit } from "@radix-ui/react-form";
import { Button, Flex, Label, Separator, Sheet, Text } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useCallback, useState } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { useNavigate, useParams } from "react-router-dom";
import { SheetHeader } from "~/components/sheet/header";
import * as z from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { CustomFieldName } from "~/components/CustomField";
import { SheetFooter } from "~/components/sheet/footer";
import { toast } from "sonner";
import { V1Beta1ServiceUserToken } from "@raystack/frontier";
import Skeleton from "react-loading-skeleton";
import styles from "../serviceusers.module.css";

const NewTokenBodySchema = z.object({
  title: z
    .string()
    .trim()
    .min(2, { message: "Must be 2 or more characters long" }),
});
export type NewTokenBody = z.infer<typeof NewTokenBodySchema>;

export default function AddServiceUserToken() {
  let { organisationId, serviceUserId } = useParams();
  const { client } = useFrontier();
  const navigate = useNavigate();

  const [token, setToken] = useState<V1Beta1ServiceUserToken>();
  const [isTokenLoading, setIsTokenLoading] = useState(false);

  const onOpenChange = useCallback(() => {
    navigate(`/organisations/${organisationId}/serviceusers/${serviceUserId}`);
    navigate(0);
  }, [navigate, organisationId, serviceUserId]);

  async function onSubmit(data: NewTokenBody) {
    setIsTokenLoading(true);
    try {
      const resp = await client?.frontierServiceCreateServiceUserToken(
        serviceUserId || "",
        data
      );

      const generatedToken = resp?.data?.token;
      setToken(generatedToken);
    } catch (err: any) {
      console.error(err);
      toast.error("Unable to create service user token", err);
    } finally {
      setIsTokenLoading(false);
    }
  }

  const methods = useForm<NewTokenBody>({
    resolver: zodResolver(NewTokenBodySchema),
    defaultValues: {},
  });

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
              title={"Generate Token"}
              onClick={onOpenChange}
            ></SheetHeader>
            <Flex
              style={{ padding: "24px" }}
              direction={"column"}
              gap={"medium"}
            >
              <CustomFieldName
                name="title"
                register={methods.register}
                control={methods.control}
              />
              <Separator />

              {isTokenLoading ? (
                <Skeleton
                  style={{
                    height: "36px",
                  }}
                />
              ) : token?.id ? (
                <Flex direction={"column"} gap={"small"}>
                  <Label size={"medium"}>ID</Label>
                  <Text size={3} className={styles.tokenTextDiv}>
                    {token?.id}
                  </Text>
                </Flex>
              ) : null}

              {isTokenLoading ? (
                <Skeleton
                  style={{
                    height: "36px",
                  }}
                />
              ) : token?.token ? (
                <Flex direction={"column"} gap={"small"}>
                  <Label size={"medium"}>Token</Label>
                  <Text size={3} className={styles.tokenTextDiv}>
                    {token?.token}
                  </Text>
                </Flex>
              ) : null}

              {token?.token ? (
                <Text
                  size={4}
                  weight={500}
                  style={{ color: "var(--foreground-danger)" }}
                >
                  For safety reasons, we can't show it again. Please copy the
                  credentials.
                </Text>
              ) : null}
            </Flex>

            <SheetFooter>
              <FormSubmit asChild>
                <Button
                  variant="primary"
                  style={{ height: "inherit" }}
                  disabled={isTokenLoading}
                >
                  <Text
                    size={4}
                    style={{ color: "var(--foreground-inverted)" }}
                  >
                    {isTokenLoading ? "Generating..." : "Generate"}
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
