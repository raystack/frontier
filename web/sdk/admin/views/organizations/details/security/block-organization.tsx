import { AlertDialog, Button, Flex, Text, toastManager } from "@raystack/apsara";
import { useContext, useState } from "react";
import { OrganizationStatus } from "../types";
import { OrganizationContext } from "../contexts/organization-context";
import { createConnectQueryKey, useMutation, useTransport } from "@connectrpc/connect-query";
import { FrontierServiceQueries, DisableOrganizationRequestSchema, EnableOrganizationRequestSchema } from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { useQueryClient } from "@tanstack/react-query";
import { useTerminology } from "../../../../hooks/useTerminology";

interface componentConfigType {
  btnColor: "danger" | "accent";
  onClick: () => void;
  btnText: string;
  dialogTitle: string;
  dialogDescription: string;
  dialogConfirmText: string;
  dialogCancelText: string;
  dialogConfirmLoadingText: string;
}

const BlockOrganizationDialog = () => {
  const t = useTerminology();
  const { organization } = useContext(OrganizationContext);
  const queryClient = useQueryClient();
  const transport = useTransport();

  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const { mutateAsync: disableOrganization, isPending: isDisabling } = useMutation(
    FrontierServiceQueries.disableOrganization,
    {
      onSuccess: async () => {
        await queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            schema: FrontierServiceQueries.getOrganization,
            transport,
            input: { id: organization?.id },
            cardinality: "finite",
          }),
        });
        setIsDialogOpen(false);
        toastManager.add({ title: `${t.organization({ case: "capital" })} blocked`, type: "success" });
      },
      onError: (error) => {
        toastManager.add({
          title: "Something went wrong",
          description: error.message,
          type: "error",
        });
        console.error("Failed to block organization:", error);
      },
    },
  );

  const { mutateAsync: enableOrganization, isPending: isEnabling } = useMutation(
    FrontierServiceQueries.enableOrganization,
    {
      onSuccess: async () => {
        await queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            schema: FrontierServiceQueries.getOrganization,
            transport,
            input: { id: organization?.id },
            cardinality: "finite",
          }),
        });
        setIsDialogOpen(false);
        toastManager.add({ title: `${t.organization({ case: "capital" })} unblocked`, type: "success" });
      },
      onError: (error) => {
        toastManager.add({
          title: "Something went wrong",
          description: error.message,
          type: "error",
        });
        console.error("Failed to unblock organization:", error);
      },
    },
  );

  async function onBlockOrganization() {
    await disableOrganization(
      create(DisableOrganizationRequestSchema, {
        id: organization?.id || "",
      }),
    );
  }

  async function onUnblockOrganization() {
    await enableOrganization(
      create(EnableOrganizationRequestSchema, {
        id: organization?.id || "",
      }),
    );
  }

  const onOpenChange = (value: boolean) => {
    setIsDialogOpen(value);
  };

  const isSubmitting = isDisabling || isEnabling;

  const componentConfig: componentConfigType =
    organization?.state === OrganizationStatus.enabled
      ? {
          btnColor: "danger",
          onClick: onBlockOrganization,
          btnText: "Block",
          dialogTitle: "Block Organization",
          dialogDescription: `Blocking this organization will restrict access to its content, disable communication, and prevent any future interactions. Are you sure you want to block ${organization?.title}?`,
          dialogConfirmText: "Block",
          dialogCancelText: "Cancel",
          dialogConfirmLoadingText: "Blocking...",
        }
      : {
          btnColor: "accent",
          onClick: onUnblockOrganization,
          btnText: "Unblock",
          dialogTitle: "Unblock Organization",
          dialogDescription: `Unblocking this organization will restore access to its content, enable communication, and allow future interactions. Are you sure you want to unblock ${organization?.title}?`,
          dialogConfirmText: "Unblock",
          dialogCancelText: "Cancel",
          dialogConfirmLoadingText: "Unblocking...",
        };

  return (
    <AlertDialog open={isDialogOpen} onOpenChange={onOpenChange}>
      <AlertDialog.Trigger
        render={
          <Button
            color={componentConfig.btnColor}
            size="small"
            data-test-id="block-orgnanization-button"
            style={{ alignSelf: "center" }}
          >
            {componentConfig.btnText}
          </Button>
        }
      />
      <AlertDialog.Content aria-label="Block Organization">
        <AlertDialog.Header>
          <AlertDialog.Title>{componentConfig.dialogTitle}</AlertDialog.Title>
          <AlertDialog.Description>
            {componentConfig.dialogDescription}
          </AlertDialog.Description>
        </AlertDialog.Header>
        <AlertDialog.Footer>
          <AlertDialog.Close
            render={
              <Button
                color="neutral"
                variant="outline"
                data-test-id="block-organization-cancel-button"
              >
                {componentConfig.dialogCancelText}
              </Button>
            }
          />
          <Button
            color={componentConfig.btnColor}
            data-test-id="block-organization-submit-button"
            loading={isSubmitting}
            loaderText={componentConfig.dialogConfirmLoadingText}
            onClick={componentConfig.onClick}
          >
            {componentConfig.dialogConfirmText}
          </Button>
        </AlertDialog.Footer>
      </AlertDialog.Content>
    </AlertDialog>
  );
};

export const BlockOrganizationSection = () => {
  return (
    <Flex gap={5} justify="between">
      <Flex direction="column" gap={3}>
        <Text size="large">Block organization</Text>
        <Text size="small" variant="secondary">
          Restrict access to safeguard platform integrity and prevent
          unauthorized activities.
        </Text>
      </Flex>
      <BlockOrganizationDialog />
    </Flex>
  );
};
