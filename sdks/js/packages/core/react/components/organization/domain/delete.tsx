import {
  Button,
  Checkbox,
  Skeleton,
  Text,
  Flex,
  toast,
  InputField,
  Dialog
} from '@raystack/apsara/v1';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useCallback, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Domain } from '~/src';
import styles from '../organization.module.css';

const domainSchema = yup
  .object({
    domain: yup
      .string()
      .required()
      .matches(/[-a-zA-Z0-9.]{1,256}\.[a-zA-Z0-9()]{1,6}$/, 'Domain is invalid')
  })
  .required();

type FormData = yup.InferType<typeof domainSchema>;

export const DeleteDomain = () => {
  const {
    watch,
    control,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
    register
  } = useForm({
    resolver: yupResolver(domainSchema)
  });
  const navigate = useNavigate({ from: '/domains/$domainId/delete' });
  const { domainId } = useParams({ from: '/domains/$domainId/delete' });
  const { client, activeOrganization: organization } = useFrontier();
  const [domain, setDomain] = useState<V1Beta1Domain>();
  const [isLoading, setIsLoading] = useState(false);
  const [isAcknowledged, setIsAcknowledged] = useState(false);

  const fetchDomainDetails = useCallback(async () => {
    if (!domainId) return;
    if (!organization?.id) return;

    try {
      setIsLoading(true);
      const res = await client?.frontierServiceGetOrganizationDomain(
        organization?.id,
        domainId
      );
      const domain = res?.data.domain;
      setDomain(domain);
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    } finally {
      setIsLoading(false);
    }
  }, [client, domainId, organization?.id]);

  useEffect(() => {
    fetchDomainDetails();
  }, [fetchDomainDetails]);

  async function onSubmit(data: FormData) {
    // @ts-ignore. TODO: fix buf openapi plugin
    if (!domain?.id || !domain?.org_id) return;

    if (data.domain !== domain.name) {
      return setError('domain', { message: 'domain name is not same' });
    }
    try {
      await client?.frontierServiceDeleteOrganizationDomain(
        // @ts-ignore
        domain.org_id,
        domain.id
      );
      navigate({ to: '/domains' });
      toast.success('Domain deleted');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  const domainName = watch('domain', '');

  return (
    <Dialog open={true}>
      <Dialog.Content width={600} overlayClassname={styles.overlay}>
        <Dialog.Header>
          <Dialog.Title>Verify domain deletion</Dialog.Title>
          <Dialog.CloseButton
            onClick={() =>
              navigate({
                to: `/domains`
              })
            }
            data-test-id="frontier-sdk-delete-domain-close-btn"
          />
        </Dialog.Header>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Body>
            <Flex direction="column" gap={5}>
              {isLoading ? (
                <>
                  <Skeleton height={'16px'} />
                  <Skeleton width={'50%'} height={'16px'} />
                  <Skeleton height={'32px'} />
                  <Skeleton height={'16px'} />
                  <Skeleton height={'32px'} />
                </>
              ) : (
                <>
                  <Text size="small">
                    This action can not be undone. This will permanently delete{' '}
                    <b>{domain?.name}</b>.
                  </Text>

                  <InputField
                    label="Please type the domain name"
                    size="large"
                    error={errors.domain && String(errors.domain?.message)}
                    {...register('domain')}
                    placeholder="Provide domain name"
                  />

                  <Flex gap="small">
                    <Checkbox
                      checked={isAcknowledged}
                      onCheckedChange={v => setIsAcknowledged(v === true)}
                      data-test-id="frontier-sdk-delete-domain-checkbox"
                    />
                    <Text size="small">
                      I acknowledge I understand that all of the team data will
                      be deleted and want to proceed.
                    </Text>
                  </Flex>

                  <Button
                    variant="solid"
                    color="danger"
                    disabled={!domainName || !isAcknowledged}
                    type="submit"
                    style={{ width: '100%' }}
                    loading={isSubmitting}
                    loaderText="Deleting..."
                    data-test-id="frontier-sdk-delete-domain-btn"
                  >
                    Delete this domain
                  </Button>
                </>
              )}
            </Flex>
          </Dialog.Body>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
