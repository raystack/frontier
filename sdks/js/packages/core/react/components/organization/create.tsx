'use client';

import { yupResolver } from '@hookform/resolvers/yup';
import { Flex, Text } from '@raystack/apsara';
import { Button, InputField } from '@raystack/apsara/v1';
import { ComponentPropsWithRef } from 'react';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { Container } from '../Container';

// @ts-ignore
import styles from './organization.module.css';

type CreateOrganizationProps = ComponentPropsWithRef<typeof Container> & {
  title?: string;
  description?: string;
};

const schema = yup
  .object({
    title: yup.string().required(),
    name: yup.string().required()
  })
  .required();

export const CreateOrganization = ({
  title = 'Create a new organization',
  description = 'Organizations are shared environments where team can work on assets, connections and data operations.',
  ...props
}: CreateOrganizationProps) => {
  const {
    handleSubmit,
    formState: { errors },
    register
  } = useForm({
    resolver: yupResolver(schema)
  });

  const { client } = useFrontier();
  async function onSubmit(data: any) {
    if (!client) return;

    const {
      data: { organization }
    } = await client.frontierServiceCreateOrganization(data);
    // @ts-ignore
    window.location = `${window.location.origin}/${organization.name}`;
  }

  return (
    <Container {...props}>
      <Flex direction="column" gap="large">
        <Flex direction="column" align="center" gap="medium">
          <Text size={9}>{title}</Text>
          <Text
            size={4}
            style={{ textAlign: 'center', color: 'var(--foreground-muted)' }}
          >
            {description}
          </Text>
        </Flex>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Container className={styles.createContainer} shadow="sm" radius="sm">
            <InputField
              label="Organization name"
              size="large"
              error={errors.title && String(errors.title?.message)}
              {...register('title')}
              placeholder="Provide organization name"
            />
            <InputField
              label="Workspace URL"
              size="large"
              error={errors.name && String(errors.name?.message)}
              {...register('name')}
              placeholder="raystack.org/"
            />
            <Button
              style={{ width: '100%' }}
              type="submit"
              data-test-id="frontier-sdk-create-workspace-btn"
            >
              Create workspace
            </Button>
          </Container>
        </form>
      </Flex>
    </Container>
  );
};
