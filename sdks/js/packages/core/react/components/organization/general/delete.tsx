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
import { useNavigate } from '@tanstack/react-router';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';

import styles from './general.module.css';
import { useTerminology } from '~/react/hooks/useTerminology';

const orgSchema = yup
  .object({
    name: yup.string()
  })
  .required();

export const DeleteOrganization = () => {
  const {
    watch,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
    register
  } = useForm({
    resolver: yupResolver(orgSchema)
  });
  const navigate = useNavigate({ from: '/delete' });
  const t = useTerminology();
  const { client, activeOrganization: organization } = useFrontier();
  const [isAcknowledged, setIsAcknowledged] = useState(false);

  async function onSubmit(data: any) {
    if (!client) return;
    if (!organization?.id) return;
    if (data.name !== organization.name)
      return setError('name', {
        message: `${t.organization({ case: 'lower' })} name is not same`
      });

    try {
      await client.frontierServiceDeleteOrganization(organization?.id);
      toast.success(`${t.organization({ case: 'capital' })} deleted`);

      // @ts-ignore
      window.location = window.location.origin;
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  const name = watch('name', '');
  return (
    <Dialog open={true}>
      <Dialog.Content overlayClassName={styles.overlay} width={600}>
        <Dialog.Header>
          <Dialog.Title>
            Verify ${t.organization({ case: 'lower' })} deletion
          </Dialog.Title>
          <Dialog.CloseButton
            onClick={() => navigate({ to: '/' })}
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
                label={`Please type name of the ${t.organization({
                  case: 'lower'
                })} to confirm.`}
                size="large"
                error={errors.name && String(errors.name?.message)}
                {...register('name')}
                placeholder={`Provide ${t.organization({
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
                I acknowledge I understand that all of the{' '}
                {t.organization({ case: 'lower' })} data will be deleted and
                want to proceed.
              </Text>
            </Flex>

            <Button
              variant="solid"
              color="danger"
              type="submit"
              disabled={!name || !isAcknowledged}
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
