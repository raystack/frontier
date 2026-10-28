'use client';

import { AlertDialog, Button, Flex, Text } from '@raystack/apsara';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useTerminology } from '../../../hooks/useTerminology';
import { blockerItems } from '../../../utils/delete-blockers';
import styles from './delete-blockers-dialog.module.css';

export type DeleteBlockersPayload = { blockers: { type: string }[] };

export interface DeleteBlockersDialogProps {
  handle: ReturnType<typeof AlertDialog.createHandle<DeleteBlockersPayload>>;
}

export function DeleteBlockersDialog({ handle }: DeleteBlockersDialogProps) {
  const t = useTerminology();
  const { basePlan } = useFrontier();
  const orgLabelLower = t.organization({ case: 'lower' });

  return (
    <AlertDialog handle={handle}>
      {({ payload: rawPayload }) => {
        const payload = rawPayload as DeleteBlockersPayload | undefined;
        const items = blockerItems(payload?.blockers ?? [], {
          organizationLabel: orgLabelLower,
          basePlanTitle: basePlan?.title
        });
        return (
          <AlertDialog.Content className={styles.content}>
            <AlertDialog.Header>
              <AlertDialog.Title>Unable to delete {orgLabelLower}</AlertDialog.Title>
            </AlertDialog.Header>
            <AlertDialog.Body>
              <Flex direction="column" gap={5}>
                <Text size="small" variant="secondary">
                  Please resolve the following issues before deleting this{' '}
                  {orgLabelLower}:
                </Text>
                <Text
                  size="small"
                  variant="secondary"
                  render={<ul className={styles.list} />}
                >
                  {items.map(({ label, instruction }) => (
                    <li key={instruction}>
                      {label ? `${label}: ${instruction}` : instruction}
                    </li>
                  ))}
                </Text>
              </Flex>
            </AlertDialog.Body>
            <AlertDialog.Footer>
              <AlertDialog.Close
                render={
                  <Button
                    variant="outline"
                    color="neutral"
                    data-test-id="frontier-sdk-delete-blockers-close-btn"
                  >
                    Close
                  </Button>
                }
              />
            </AlertDialog.Footer>
          </AlertDialog.Content>
        );
      }}
    </AlertDialog>
  );
}
