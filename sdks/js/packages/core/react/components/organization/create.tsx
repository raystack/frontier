'use client';

import { yupResolver } from '@hookform/resolvers/yup';
import { Button, Flex, InputField, Text, TextField } from '@raystack/apsara';
import { ComponentPropsWithRef } from 'react';
import { Controller, useForm } from 'react-hook-form';
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
    control,
    register,
    handleSubmit,
    formState: { errors }
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
            <InputField label="Organization name">
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    // @ts-ignore
                    size="medium"
                    placeholder="Provide organization name"
                  />
                )}
                control={control}
                name="title"
              />

              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.title && String(errors.title?.message)}
              </Text>
            </InputField>
            <InputField label="Workspace URL">
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    // @ts-ignore
                    size="medium"
                    label="Workspace URL"
                    placeholder="raystack.org/"
                  />
                )}
                control={control}
                name="name"
              />
              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.name && String(errors.name?.message)}
              </Text>
            </InputField>

            <Button
              variant="primary"
              size="medium"
              style={{ width: '100%' }}
              type="submit"
            >
              Create workspace
            </Button>
          </Container>
        </form>
      </Flex>
    </Container>
  );
};
