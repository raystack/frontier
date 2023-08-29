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

const domainSchema = yup
  .object({
    domain: yup.string().required()
  })
  .required();

export const AddDomain = ({
  organization
}: {
  organization?: V1Beta1Organization;
}) => {
  const {
    reset,
    control,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(domainSchema)
  });
  const navigate = useNavigate();
  const { client } = useFrontier();

  async function onSubmit(data: any) {
    if (!client) return;
    if (!organization?.id) return;

    try {
      await client.frontierServiceCreateOrganizationDomain(
        organization?.id,
        data
      );
      toast.success('Domain added');

      navigate('/domains');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return (
    <Dialog open={true}>
      <Dialog.Content style={{ padding: 0, maxWidth: '600px', width: '100%' }}>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex justify="between" style={{ padding: '16px 24px' }}>
            <Text size={6} style={{ fontWeight: '500' }}>
              Add domain
            </Text>

            <Image
              alt="cross"
              // @ts-ignore
              src={cross}
              onClick={() => navigate('/domains')}
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
