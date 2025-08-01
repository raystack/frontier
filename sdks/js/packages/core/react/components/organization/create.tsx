'use client';

import { yupResolver } from '@hookform/resolvers/yup';
import { Button, Text, Headline, Flex, InputField } from '@raystack/apsara/v1';
import { ComponentPropsWithRef } from 'react';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { Container } from '../Container';

// @ts-ignore
import styles from './organization.module.css';
import { useTerminology } from '~/react/hooks/useTerminology';

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
  const t = useTerminology();
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
      <Flex direction="column" gap={9}>
        <Flex direction="column" align="center" gap={5}>
          <Headline size="t2">{title}</Headline>
          <Text
            size="regular"
            variant="secondary"
            style={{ textAlign: 'center' }}
          >
            {description}
          </Text>
        </Flex>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Container className={styles.createContainer} shadow="sm" radius="sm">
            <InputField
              label={`${t.organization({ case: 'capital' })} name`}
              size="large"
              error={errors.title && String(errors.title?.message)}
              {...register('title')}
              placeholder={`Provide ${t.organization({ case: 'lower' })} name`}
            />
            <InputField
              label={`${t.organization({ case: 'capital' })} URL`}
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
              Create {t.organization({ case: 'lower' })}
            </Button>
          </Container>
        </form>
      </Flex>
    </Container>
  );
};
