'use client';

import { useCallback, useState } from 'react';
import { create } from '@bufbuild/protobuf';
import { timestampFromDate } from '@bufbuild/protobuf/wkt';
import { useMutation } from '@connectrpc/connect-query';
import dayjs from 'dayjs';
import {
  FrontierServiceQueries,
  RegenerateCurrentUserPATRequestSchema
} from '@raystack/proton/frontier';
import {
  Button,
  Dialog,
  Flex,
  Label,
  Select,
  Text,
  toastManager
} from '@raystack/apsara-v1';
import { useFrontier } from '../../../contexts/FrontierContext';
import { DEFAULT_DATE_FORMAT } from '../../../utils/constants';
import { handleConnectError } from '~/utils/error';

const EXPIRY_OPTIONS = [15, 30, 60, 90] as const;

export interface RegeneratePayload {
  patId: string;
  currentExpiryDays: string;
}

export interface RegeneratePATDialogProps {
  handle: ReturnType<typeof Dialog.createHandle<RegeneratePayload>>;
  onRegenerated?: (token: string) => void;
}

export function RegeneratePATDialog({
  handle,
  onRegenerated
}: RegeneratePATDialogProps) {
  const { config } = useFrontier();
  const dateFormat = config?.dateFormat || DEFAULT_DATE_FORMAT;

  const [expiryDays, setExpiryDays] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { mutateAsync: regeneratePAT } = useMutation(
    FrontierServiceQueries.regenerateCurrentUserPAT
  );

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setExpiryDays('');
    }
  };

  const handleRegenerate = useCallback(async () => {
    const days = expiryDays || handle.payload?.currentExpiryDays;
    if (!days) return;

    setIsSubmitting(true);
    try {
      const expiresAt = timestampFromDate(
        dayjs().add(Number(days), 'day').toDate()
      );

      const patId = handle.payload?.patId;
      if (!patId) return;

      const response = await regeneratePAT(
        create(RegenerateCurrentUserPATRequestSchema, {
          id: patId,
          expiresAt
        })
      );

      const token = response.pat?.token;
      toastManager.add({
        title: 'Token regenerated',
        type: 'success'
      });
      handle.close();
      setExpiryDays('');
      if (token) onRegenerated?.(token);
    } catch (error) {
      handleConnectError(error, {
        Default: err =>
          toastManager.add({
            title: 'Something went wrong',
            description: err.message,
            type: 'error'
          })
      });
    } finally {
      setIsSubmitting(false);
    }
  }, [expiryDays, handle, regeneratePAT, onRegenerated]);

  return (
    <Dialog handle={handle} onOpenChange={handleOpenChange}>
      {({ payload }) => {
        const selectedDays = expiryDays || payload?.currentExpiryDays || '';
        return (
          <Dialog.Content width={400}>
            <Dialog.Header>
              <Dialog.Title>Regenerate Expiry date</Dialog.Title>
            </Dialog.Header>
            <Dialog.Body>
              <Flex direction="column" gap={7}>
                <Text size="small">
                  Select a new expiry duration for this personal access token.
                  The current token will be invalidated and a new one will be
                  generated.
                </Text>
                <Flex direction="column" gap={2}>
                  <Label>Expiry date</Label>
                  <Select value={selectedDays} onValueChange={setExpiryDays}>
                    <Select.Trigger>
                      <Select.Value placeholder="Select expiry" />
                    </Select.Trigger>
                    <Select.Content>
                      {EXPIRY_OPTIONS.map(days => (
                        <Select.Item key={days} value={String(days)}>
                          {days} Days (Exp:{' '}
                          {dayjs().add(days, 'day').format(dateFormat)})
                        </Select.Item>
                      ))}
                    </Select.Content>
                  </Select>
                </Flex>
              </Flex>
            </Dialog.Body>
            <Dialog.Footer>
              <Flex justify="end">
                <Button
                  variant="solid"
                  color="accent"
                  size="normal"
                  onClick={handleRegenerate}
                  loading={isSubmitting}
                  disabled={!selectedDays || isSubmitting}
                  loaderText="Regenerating..."
                  data-test-id="frontier-sdk-pat-regenerate-submit-btn"
                >
                  Regenerate
                </Button>
              </Flex>
            </Dialog.Footer>
          </Dialog.Content>
        );
      }}
    </Dialog>
  );
}
