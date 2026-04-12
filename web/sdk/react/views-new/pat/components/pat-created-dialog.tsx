'use client';

import { InfoCircledIcon } from '@radix-ui/react-icons';
import {
  Button,
  Callout,
  CopyButton,
  Dialog,
  Flex,
  InputField,
  Text
} from '@raystack/apsara-v1';
import styles from '../pat-view.module.css';

export interface PATCreatedDialogProps {
  handle: ReturnType<typeof Dialog.createHandle<string>>;
  onClose?: () => void;
}

export function PATCreatedDialog({ handle, onClose }: PATCreatedDialogProps) {
  const handleOpenChange = (open: boolean) => {
    if (!open) onClose?.();
  };

  return (
    <Dialog handle={handle} onOpenChange={handleOpenChange}>
      {({ payload: token }) => (
        <Dialog.Content width={400}>
          <Dialog.Header>
            <Dialog.Title>Success</Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Flex direction="column" gap={7}>
              <Text size="small">
                You&apos;ve successfully added a new personal access token. Copy
                the token now
              </Text>
              <InputField
                value={token || ''}
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
              <Callout type="alert" icon={<InfoCircledIcon />} className={styles.callout}>
                Warning : Make sure you copy the above token now. We don&apos;t
                store it and you will not be able to see it again.
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
      )}
    </Dialog>
  );
}
