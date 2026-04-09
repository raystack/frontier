import { useState } from 'react';
import { yupResolver } from '@hookform/resolvers/yup';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { create } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  DeleteOrganizationRequestSchema
} from '@raystack/proton/frontier';
import {
  Button,
  Checkbox,
  Text,
  Flex,
  Dialog,
  InputField,
  toastManager
} from '@raystack/apsara-v1';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useTerminology } from '../../../hooks/useTerminology';
import styles from './delete-organization-dialog.module.css';
import { handleConnectError } from '~/utils/error';

const deleteOrgSchema = yup
  .object({
    title: yup.string()
  })
  .required();

export interface DeleteOrganizationDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onDeleteSuccess?: () => void;
}

export const DeleteOrganizationDialog = ({
  open,
  onOpenChange,
  onDeleteSuccess
}: DeleteOrganizationDialogProps) => {
  const t = useTerminology();
  const { activeOrganization: organization } = useFrontier();
  const orgLabel = t.organization({ case: 'capital' });
  const orgLabelLower = t.organization({ case: 'lower' });
  const [isAcknowledged, setIsAcknowledged] = useState(false);

  const { mutateAsync: deleteOrganization } = useMutation(
    FrontierServiceQueries.deleteOrganization
  );

  const {
    register,
    handleSubmit,
    watch,
    setError,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(deleteOrgSchema)
  });

  const deleteTitle = watch('title') ?? '';

  async function onDeleteSubmit(data: { title?: string }) {
    if (!organization?.id) return;
    if (data.title !== organization.title) {
      setError('title', {
        message: `The ${orgLabelLower} name does not match`
      });
      return;
    }

    try {
      const req = create(DeleteOrganizationRequestSchema, {
        id: organization.id
      });
      await deleteOrganization(req);
      toastManager.add({
        title: `${orgLabel} deleted`,
        type: 'success'
      });
      onDeleteSuccess?.();
    } catch (error) {
      handleConnectError(error, {
        PermissionDenied: () => toastManager.add({ title: "You don't have permission to perform this action", type: 'error' }),
        NotFound: (err) => toastManager.add({ title: 'Not found', description: err.message, type: 'error' }),
        Default: (err) => toastManager.add({ title: 'Something went wrong', description: err.message, type: 'error' }),
      });
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <Dialog.Content width={432}>
        <Dialog.Header>
          <Dialog.Title>
            Verify {orgLabel} deletion
          </Dialog.Title>
        </Dialog.Header>
        <form onSubmit={handleSubmit(onDeleteSubmit)}>
          <Dialog.Body>
            <Flex direction="column" gap={5}>
              <Text size="small" variant="secondary">
                This action can not be undone. This will permanently
                delete all the projects and resources in {organization?.title}.
              </Text>
              <InputField
                label={`Please type name of the ${orgLabel} to confirm.`}
                size="large"
                error={
                  errors.title
                    ? String(errors.title.message)
                    : undefined
                }
                {...register('title')}
                placeholder={`Provide the ${orgLabel} name`}
              />
              <Flex gap={3} align="start">
                <Checkbox
                  checked={isAcknowledged}
                  onCheckedChange={(checked: boolean) => setIsAcknowledged(checked)}
                  data-test-id="frontier-sdk-delete-organization-checkbox"
                />
                <Text size="small" variant="secondary">
                  I acknowledge I understand that all of the{' '}
                  {orgLabel} data will be deleted and want to proceed.
                </Text>
              </Flex>
              <Button
                variant="solid"
                color="danger"
                type="submit"
                disabled={!deleteTitle || !isAcknowledged}
                className={styles.deleteButton}
                data-test-id="frontier-sdk-delete-organization-btn"
                loading={isSubmitting}
                loaderText="Deleting..."
              >
                Delete this {orgLabelLower}
              </Button>
            </Flex>
          </Dialog.Body>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
