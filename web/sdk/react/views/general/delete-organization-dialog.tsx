import { useState } from 'react';
import {
  Button,
  Checkbox,
  toast,
  Text,
  Flex,
  Dialog,
  InputField
} from '@raystack/apsara';

import { yupResolver } from '@hookform/resolvers/yup';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  DeleteOrganizationRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { useTerminology } from '~/react/hooks/useTerminology';

import styles from './general.module.css';

const orgSchema = yup
  .object({
    title: yup.string()
  })
  .required();

export interface DeleteOrganizationDialogProps {
  open: boolean;
  onOpenChange: (value: boolean) => void;
  onDeleteSuccess?: () => void;
}

export const DeleteOrganizationDialog = ({
  open,
  onOpenChange,
  onDeleteSuccess
}: DeleteOrganizationDialogProps) => {
  const {
    watch,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
    register
  } = useForm({
    resolver: yupResolver(orgSchema)
  });
  const t = useTerminology();
  const { activeOrganization: organization } = useFrontier();
  const { mutateAsync: deleteOrganization } = useMutation(
    FrontierServiceQueries.deleteOrganization
  );
  const [isAcknowledged, setIsAcknowledged] = useState(false);

  async function onSubmit(data: any) {
    if (!organization?.id) return;
    if (data.title !== organization.title)
      return setError('title', {
        message: `The ${t.organization({ case: 'lower' })} name does not match`
      });

    try {
      const req = create(DeleteOrganizationRequestSchema, {
        id: organization.id
      });
      await deleteOrganization(req);
      toast.success(`${t.organization({ case: 'capital' })} deleted`);

      onDeleteSuccess?.();
    } catch (error: any) {
      toast.error('Something went wrong', {
        description:
          error?.message ||
          `Failed to delete ${t.organization({ case: 'capital' })}`
      });
    }
  }

  const title = watch('title', '');
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <Dialog.Content overlayClassName={styles.overlay} width={600}>
        <Dialog.Header>
          <Dialog.Title>
            Verify {t.organization({ case: 'lower' })} deletion
          </Dialog.Title>
          <Dialog.CloseButton
            data-test-id="frontier-sdk-delete-organization-close-btn"
          />
        </Dialog.Header>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Body>
            <Flex direction="column" gap={5}>
              <Text size="small">
                This action <b>can not</b> be undone. This will permanently
                delete all the projects and resources in{' '}
                <b>{organization?.title}</b>.
              </Text>
              <InputField
                label={`Please enter the name of the ${t.organization({
                  case: 'lower'
                })} to confirm.`}
                size="large"
                error={errors.title && String(errors.title?.message)}
                {...register('title')}
                placeholder={`Provide the ${t.organization({
                  case: 'lower'
                })} name`}
              />
            </Flex>
          </Dialog.Body>
          <Dialog.Footer className={styles.deleteFooter}>
            <Flex gap={3}>
              <Checkbox
                checked={isAcknowledged}
                onCheckedChange={v => setIsAcknowledged(v === true)}
                data-test-id="frontier-sdk-delete-organization-checkbox"
              />
              <Text size="small">
                I acknowledge and understand that all of the{' '}
                {t.organization({ case: 'lower' })} data will be deleted and
                want to proceed.
              </Text>
            </Flex>

            <Button
              variant="solid"
              color="danger"
              type="submit"
              disabled={!title || !isAcknowledged}
              style={{ width: '100%' }}
              data-test-id="frontier-sdk-delete-organization-btn"
              loading={isSubmitting}
              loaderText="Deleting..."
            >
              Delete this {t.organization({ case: 'lower' })}
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};

