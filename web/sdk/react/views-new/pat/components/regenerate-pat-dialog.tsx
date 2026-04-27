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
  Select,
  Text,
  toastManager
} from '@raystack/apsara-v1';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import { handleConnectError } from '~/utils/error';
import { EXPIRY_OPTIONS } from '../utils';
import styles from './regenerate-pat-dialog.module.css';

export interface RegeneratePayload {
  patId: string;
  currentExpiryValue: string;
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

  const [expiryValue, setExpiryValue] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { mutateAsync: regeneratePAT } = useMutation(
    FrontierServiceQueries.regenerateCurrentUserPAT
  );

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setExpiryValue('');
    }
  };

  const handleRegenerate = useCallback(
    async (patId: string, selectedValue: string) => {
      const option = EXPIRY_OPTIONS.find(o => o.value === selectedValue);
      if (!option) return;

      setIsSubmitting(true);
      try {
        const expiresAt = timestampFromDate(
          dayjs().add(option.amount, option.unit).toDate()
        );

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
        setExpiryValue('');
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
    },
    [regeneratePAT, handle, onRegenerated]
  );

  return (
    <Dialog handle={handle} onOpenChange={handleOpenChange}>
      {({ payload }) => {
        const selectedValue =
          expiryValue || payload?.currentExpiryValue || '';
        const patId = payload?.patId;
        return (
          <Dialog.Content width={400}>
            <Dialog.Header>
              <Dialog.Title>Regenerate Expiry date</Dialog.Title>
            </Dialog.Header>
            <Dialog.Body className={styles.body}>
              <Text size="small">
                Select a new expiry duration for this personal access token.
                The current token will be invalidated and replaced with a new
                one.
              </Text>
              <Select value={selectedValue} onValueChange={setExpiryValue}>
                <Select.Trigger>
                  <Select.Value placeholder="Select expiry" />
                </Select.Trigger>
                <Select.Content>
                  {EXPIRY_OPTIONS.map(option => (
                    <Select.Item key={option.value} value={option.value}>
                      {option.label} (Exp:{' '}
                      {dayjs()
                        .add(option.amount, option.unit)
                        .format(dateFormat)}
                      )
                    </Select.Item>
                  ))}
                </Select.Content>
              </Select>
            </Dialog.Body>
            <Dialog.Footer>
              <Flex justify="end">
                <Button
                  variant="solid"
                  color="accent"
                  size="normal"
                  onClick={() =>
                    patId && handleRegenerate(patId, selectedValue)
                  }
                  loading={isSubmitting}
                  disabled={!selectedValue || isSubmitting}
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
