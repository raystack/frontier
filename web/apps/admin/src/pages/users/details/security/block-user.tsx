import { ComponentProps, useCallback, useState } from "react";
import { Button, Dialog, toast } from "@raystack/apsara";
import { useUser } from "../user-context";
import { getUserName } from "../../util";
import {
  createConnectQueryKey,
  useMutation,
  useTransport,
} from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  FrontierServiceQueries,
} from "@raystack/proton/frontier";
import { useQueryClient } from "@tanstack/react-query";

type ButtonColorType = ComponentProps<typeof Button>["color"];

export const BlockUserDialog = () => {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const { user, reset } = useUser();
  const queryClient = useQueryClient();
  const transport = useTransport();

  const isActive = user?.state === "enabled";
  const userName = getUserName(user);

  const optimisticUpdateState = useCallback(
    (state: string) => {
      queryClient.setQueryData(
        createConnectQueryKey({
          schema: AdminServiceQueries.searchUsers,
          transport,
          input: { query: { search: user?.id } },
          cardinality: "finite",
        }),
        oldData => {
          if (!oldData) return oldData;
          return {
            ...oldData,
            users: oldData.users.map(user =>
              user.id === user?.id ? { ...user, state } : user,
            ),
          };
        },
      );
    },
    [queryClient, transport, user?.id],
  );

  const { mutateAsync: blockUser, isPending: isBlockingUser } = useMutation(
    FrontierServiceQueries.disableUser,
    {
      onSuccess: () => {
        toast.success("User blocked successfully");
        optimisticUpdateState("disabled");
        reset?.();
        onOpenChange(false);
      },
      onError: error => {
        console.error("Failed to block user", error);
      },
    },
  );
  const { mutateAsync: unblockUser, isPending: isUnblockingUser } = useMutation(
    FrontierServiceQueries.enableUser,
    {
      onSuccess: () => {
        toast.success("User unblocked successfully");
        optimisticUpdateState("enabled");
        reset?.();
        onOpenChange(false);
      },
      onError: error => {
        console.error("Failed to block user", error);
      },
    },
  );

  const isSubmitting = isBlockingUser || isUnblockingUser;

  const onOpenChange = (value: boolean) => {
    setIsDialogOpen(value);
  };

  const config = isActive
    ? {
        onClick: blockUser,
        btnColor: "danger" as ButtonColorType,
        btnText: "Block",
        dialogTitle: "Block user",
        dialogDescription: `Blocking this user will permanently restrict access to its content, disable communication, and prevent any future interactions. Are you sure you want to suspend ${userName}?`,
        dialogConfirmText: "Block",
        dialogConfirmLoadingText: "Blocking...",
      }
    : {
        onClick: unblockUser,
        btnColor: "accent" as ButtonColorType,
        btnText: "Unblock",
        dialogTitle: "Unblock user",
        dialogDescription: `Unblocking this user will restore access to its content, enable communication, and allow future interactions. Are you sure you want to unblock ${userName}?`,
        dialogConfirmText: "Unblock",
        dialogConfirmLoadingText: "Unblocking...",
      };

  return (
    <Dialog open={isDialogOpen} onOpenChange={onOpenChange}>
      <Dialog.Trigger asChild>
        <Button
          color={config.btnColor}
          size="normal"
          style={{
            alignSelf: "flex-end",
          }}
          width={70}
          data-test-id="admin-security-block-user">
          {config.btnText}
        </Button>
      </Dialog.Trigger>
      <Dialog.Content width={400} ariaLabel="Block user">
        <Dialog.Body>
          <Dialog.Title>{config.dialogTitle}</Dialog.Title>
          <Dialog.Description>{config.dialogDescription}</Dialog.Description>
        </Dialog.Body>
        <Dialog.Footer>
          <Dialog.Close asChild>
            <Button
              color="neutral"
              variant="outline"
              data-test-id="admin-security-block-user-cancel">
              Cancel
            </Button>
          </Dialog.Close>
          <Button
            color={config.btnColor}
            data-test-id="admin-security-block-user-submit"
            loading={isSubmitting}
            loaderText={config.dialogConfirmLoadingText}
            onClick={() => config.onClick({ id: user?.id || "" })}>
            {config.dialogConfirmText}
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
