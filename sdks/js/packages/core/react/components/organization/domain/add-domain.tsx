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
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';

const domainSchema = yup
  .object({
    domain: yup.string().required()
  })
  .required();

export const AddDomain = () => {
  const {
    control,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(domainSchema)
  });
  const navigate = useNavigate({ from: '/domains/modal' });
  const { client, activeOrganization: organization } = useFrontier();

  async function onSubmit(data: any) {
    if (!client) return;
    if (!organization?.id) return;

    try {
      const {
        data: { domain }
      } = await client.frontierServiceCreateOrganizationDomain(
        organization?.id,
        data
      );
      toast.success('Domain added');

      navigate({ to: '/domains' });
      navigate({
        to: `/domains/$domainId/verify`,
        params: { domainId: domain?.id }
      });
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content style={{ padding: 0, maxWidth: '600px', width: '100%' }}>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex justify="between" style={{ padding: '16px 24px' }}>
            <Text size={6} style={{ fontWeight: '500' }}>
              Add domain
            </Text>

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              // @ts-ignore
              src={cross}
              onClick={() => navigate({ to: '/domains' })}
            />
          </Flex>
          <Separator />

          <Flex
            direction="column"
            gap="medium"
            style={{ padding: '24px 32px' }}
          >
            <InputField label="Domain name">
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    // @ts-ignore
                    size="medium"
                    placeholder="Provide domain name"
                  />
                )}
                control={control}
                name="domain"
              />

              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.domain && String(errors.domain?.message)}
              </Text>
            </InputField>
          </Flex>
          <Separator />
          <Flex justify="end" style={{ padding: 'var(--pd-16)' }}>
            <Button variant="primary" size="medium" type="submit">
              {isSubmitting ? 'adding...' : 'Add domain'}
            </Button>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
