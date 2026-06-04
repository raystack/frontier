'use client';

import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import {
  Button,
  Callout,
  CopyButton,
  Dialog,
  Flex,
  Input,
  Text
} from '@raystack/apsara';

export interface PATCreatedPayload {
  token: string;
  isRegenerated?: boolean;
}

export interface PATCreatedDialogProps {
  handle: ReturnType<typeof Dialog.createHandle<PATCreatedPayload>>;
  onClose?: () => void;
}

export function PATCreatedDialog({ handle, onClose }: PATCreatedDialogProps) {
  const handleOpenChange = (open: boolean) => {
    if (!open) onClose?.();
  };

  return (
    <Dialog handle={handle} onOpenChange={handleOpenChange}>
      {({ payload }) => {
        const token = payload?.token ?? '';
        const isRegenerated = payload?.isRegenerated ?? false;
        const description = isRegenerated
          ? 'Your personal access token has been regenerated successfully. Please copy and store it securely.'
          : 'Successfully added a new personal access token. Please copy the token.';
        return (
          <Dialog.Content>
            <Dialog.Header>
              <Dialog.Title>Success</Dialog.Title>
            </Dialog.Header>
            <Dialog.Body>
              <Flex direction="column" gap={7}>
                <Text size="small">{description}</Text>
                <Input
                  value={token}
                  readOnly
                  trailingIcon={
                    token ? (
                      <CopyButton
                        text={token}
                        size={2}
                        data-test-id="frontier-sdk-pat-token-copy-btn"
                      />
                    ) : undefined
                  }
                  data-test-id="frontier-sdk-pat-token-input"
                />
                <Callout
                  type="attention"
                  outline
                  icon={<ExclamationTriangleIcon />}
                  width="100%"
                >
                  Warning: Make sure you copy the above token now. This token
                  will only be shown once. Store it securely.
                </Callout>
              </Flex>
            </Dialog.Body>
            <Dialog.Footer>
              <Flex justify="end">
                <Button
                  variant="solid"
                  color="accent"
                  size="normal"
                  onClick={() => handle.close()}
                  data-test-id="frontier-sdk-pat-created-close-btn"
                >
                  Close
                </Button>
              </Flex>
            </Dialog.Footer>
          </Dialog.Content>
        );
      }}
    </Dialog>
  );
}
