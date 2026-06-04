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
  AlertDialog,
  Button,
  Checkbox,
  Field,
  Flex,
  Input,
  Text,
  toastManager
} from '@raystack/apsara';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useTerminology } from '../../../hooks/useTerminology';
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
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialog.Content width={480}>
        <form onSubmit={handleSubmit(onDeleteSubmit)}>
          <AlertDialog.Header>
            <AlertDialog.Title>Delete {orgLabel}</AlertDialog.Title>
          </AlertDialog.Header>
          <AlertDialog.Body>
            <Flex direction="column" gap={6}>
              <Text size="small" variant="secondary">
                This action can not be undone. This will permanently
                delete all the projects and resources in {organization?.title}.
              </Text>
              <Field
                label={`Please type name of the ${orgLabel} to confirm.`}
                error={
                  errors.title
                    ? String(errors.title.message)
                    : undefined
                }
              >
                <Input
                  size="large"
                  {...register('title')}
                  placeholder={`Provide the ${orgLabel} name`}
                />
              </Field>
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
            </Flex>
          </AlertDialog.Body>
          <AlertDialog.Footer>
            <AlertDialog.Close
              render={
                <Button
                  variant="outline"
                  color="neutral"
                  disabled={isSubmitting}
                  data-test-id="frontier-sdk-delete-organization-cancel-btn"
                >
                  Cancel
                </Button>
              }
            />
            <Button
              variant="solid"
              color="danger"
              type="submit"
              disabled={!deleteTitle || !isAcknowledged}
              data-test-id="frontier-sdk-delete-organization-btn"
              loading={isSubmitting}
              loaderText="Deleting..."
            >
              Delete
            </Button>
          </AlertDialog.Footer>
        </form>
      </AlertDialog.Content>
    </AlertDialog>
  );
};
