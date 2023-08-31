import {
  Button,
  Dialog,
  Flex,
  Image,
  InputField,
  Separator,
  Text,
  TextField
} from '@raystack/apsara';

import { yupResolver } from '@hookform/resolvers/yup';
import { Controller, useForm } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Organization } from '~/src';

// @ts-ignore
import styles from "./general.module.css";

const orgSchema = yup
  .object({
    name: yup.string()
  })
  .required();

export const DeleteOrganization = ({
  organization
}: {
  organization?: V1Beta1Organization;
}) => {
  const {
    watch,
    control,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(orgSchema)
  });
  const navigate = useNavigate();
  const { client } = useFrontier();

  async function onSubmit(data: any) {
    if (!client) return;
    if (!organization?.id) return;
    if (data.name !== organization.name)
      return setError('name', { message: 'organization name is not same' });

    try {
      await client.frontierServiceDeleteOrganization(organization?.id);
      toast.success('Organization deleted');

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
      <Dialog.Content style={{ padding: 0, maxWidth: '600px', width: '100%' }}>
        <Flex justify="between" style={{ padding: '16px 24px' }}>
          <Text size={6} style={{ fontWeight: '500' }}>
            Verify organisation deletion
          </Text>

          <Image
            className={styles.deleteIcon}
            alt="cross"
            // @ts-ignore
            src={cross}
            onClick={() => navigate('/')}
          />
        </Flex>
        <Separator />
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex
            direction="column"
            gap="medium"
            style={{ padding: '24px 32px' }}
          >
            <Text size={2}>
              This action <b>can not</b> be undone. This will permanently delete
              all the projects and resources in <b>{organization?.title}</b>.
            </Text>

            <InputField label="Please type name of the organisation to confirm.">
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    // @ts-ignore
                    size="medium"
                    placeholder="Provide organisation name"
                  />
                )}
                control={control}
                name="name"
              />

              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.name && String(errors.name?.message)}
              </Text>
            </InputField>
            <Flex>
              <Text size={2}>
                I acknowledge I understand that all of the organisation data
                will be deleted and want to proceed.
              </Text>
            </Flex>

            <Button
              variant="danger"
              size="medium"
              type="submit"
              disabled={!name}
              style={{ width: '100%' }}
            >
              {isSubmitting ? 'deleting...' : 'Delete this organization'}
            </Button>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
