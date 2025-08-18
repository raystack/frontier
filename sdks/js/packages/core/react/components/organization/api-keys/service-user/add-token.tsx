import * as yup from 'yup';
import { yupResolver } from '@hookform/resolvers/yup';
import { useForm } from 'react-hook-form';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useCallback } from 'react';
import { Flex, toast, Button, InputField } from '@raystack/apsara';
import { V1Beta1ServiceUserToken } from '~/api-client';
import styles from './styles.module.css';

const serviceAccountSchema = yup
  .object({
    title: yup.string().required('Name is a required field')
  })
  .required();

type FormData = yup.InferType<typeof serviceAccountSchema>;

export default function AddServiceUserToken({
  serviceUserId,
  onAddToken = () => {}
}: {
  serviceUserId: string;
  onAddToken: (token: V1Beta1ServiceUserToken) => void;
}) {
  const { client, activeOrganization } = useFrontier();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(serviceAccountSchema)
  });

  const orgId = activeOrganization?.id || '';

  const onSubmit = useCallback(
    async (data: FormData) => {
      if (!client) return;

      try {
        const {
          data: { token }
        } = await client.frontierServiceCreateServiceUserToken(
          orgId,
          serviceUserId,
          data
        );
        if (token) {
          onAddToken(token);
          toast.success('Api key created');
        }
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    },
    [client, onAddToken, serviceUserId, orgId]
  );

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Flex gap={3}>
        <Flex className={styles.addKeyInputWrapper} gap={3}>
          <InputField
            {...register('title')}
            size="large"
            placeholder="Provide service key name"
            error={errors.title && String(errors.title?.message)}
          />
          <Button
            data-test-id="frontier-sdk-api-keys-new-token-btn"
            type="submit"
            loading={isSubmitting}
            disabled={isSubmitting}
            loaderText="Generating..."
          >
            Generate new key
          </Button>
        </Flex>
      </Flex>
    </form>
  );
}
