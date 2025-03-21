import { Button, Dialog, Flex, Text, toast } from "@raystack/apsara/v1";
import { useState } from "react";
import { useOutletContext } from "react-router-dom";
import { api } from "~/api";
import { OutletContext, OrganizationStatus } from "../types";

const BlockOrganizationDialog = () => {
  const { organization, fetchOrganization } = useOutletContext<OutletContext>();

  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function onBlockOrganization() {
    try {
      setIsSubmitting(true);
      await api?.frontierServiceDisableOrganization(organization.id || "", {});
      await fetchOrganization(organization.id || "");
      onOpenChange(false);
    } catch (error) {
      console.error("Failed to block organization", error);
      toast.error("Failed to block organization");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function onUnblockOrganization() {
    try {
      setIsSubmitting(true);
      await api?.frontierServiceEnableOrganization(organization.id || "", {});
      await fetchOrganization(organization.id || "");
      onOpenChange(false);
    } catch (error) {
      console.error("Failed to unblock organization", error);
      toast.error("Failed to unblock organization");
    } finally {
      setIsSubmitting(false);
    }
  }

  const onOpenChange = (value: boolean) => {
    setIsDialogOpen(value);
  };

  const componentConfig =
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
          btnColor: "primary",
          onClick: onUnblockOrganization,
          btnText: "Unblock",
          dialogTitle: "Unblock Organization",
          dialogDescription: `Unblocking this organization will restore access to its content, enable communication, and allow future interactions. Are you sure you want to unblock ${organization?.title}?`,
          dialogConfirmText: "Unblock",
          dialogCancelText: "Cancel",
          dialogConfirmLoadingText: "Unblocking...",
        };

  return (
    <Dialog open={isDialogOpen} onOpenChange={onOpenChange}>
      <Dialog.Trigger asChild>
        <Button
          color={componentConfig.btnColor}
          size="small"
          data-test-id="block-orgnanization-button"
        >
          {componentConfig.btnText}
        </Button>
      </Dialog.Trigger>
      <Dialog.Content width={400} ariaLabel="Block Organization">
        <Dialog.Body>
          <Dialog.Title>{componentConfig.dialogTitle}</Dialog.Title>
          <Dialog.Description>
            {componentConfig.dialogDescription}
          </Dialog.Description>
        </Dialog.Body>
        <Dialog.Footer>
          <Dialog.Close asChild>
            <Button
              color="neutral"
              variant="outline"
              data-test-id="block-organization-cancel-button"
            >
              {componentConfig.dialogCancelText}
            </Button>
          </Dialog.Close>
          <Button
            color={componentConfig.btnColor}
            data-test-id="block-organization-submit-button"
            loading={isSubmitting}
            loaderText={componentConfig.dialogConfirmLoadingText}
            onClick={componentConfig.onClick}
          >
            {componentConfig.dialogConfirmText}
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

export const BlockOrganizationSection = () => {
  return (
    <Flex gap={5} justify="between">
      <Flex direction="column" gap={3}>
        <Text size={5}>Block organization</Text>
        <Text size={3} variant="secondary">
          Restrict access to safeguard platform integrity and prevent
          unauthorized activities.
        </Text>
      </Flex>
      <BlockOrganizationDialog />
    </Flex>
  );
};
