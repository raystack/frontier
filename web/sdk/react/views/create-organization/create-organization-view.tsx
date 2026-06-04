'use client';

import { yupResolver } from '@hookform/resolvers/yup';
import {
  Button,
  Text,
  Headline,
  Flex,
  Field,
  Input,
  toastManager
} from '@raystack/apsara';
import { ComponentPropsWithRef } from 'react';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  CreateOrganizationRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { useTerminology } from '~/react/hooks/useTerminology';
import styles from './create-organization-view.module.css';

export type CreateOrganizationViewProps = ComponentPropsWithRef<'div'> & {
  title?: string;
  description?: string;
};

const schema = yup
  .object({
    title: yup.string().required(),
    name: yup.string().required()
  })
  .required();

type FormData = yup.InferType<typeof schema>;

export const CreateOrganizationView = ({
  title = 'Create a new organization',
  description = 'Organizations are shared environments where team can work on assets, connections and data operations.',
  ...props
}: CreateOrganizationViewProps) => {
  const t = useTerminology();
  const {
    handleSubmit,
    formState: { errors, isSubmitting },
    register
  } = useForm<FormData>({
    resolver: yupResolver(schema)
  });

  const { mutateAsync: createOrganization } = useMutation(
    FrontierServiceQueries.createOrganization,
    {
      onError: (err: Error) => {
        toastManager.add({
          title: 'Failed to create organization',
          description: err?.message,
          type: 'error'
        });
      }
    }
  );

  async function onSubmit(data: FormData) {
    const response = await createOrganization(
      create(CreateOrganizationRequestSchema, {
        body: {
          title: data.title,
          name: data.name
        }
      })
    );
    const organization = response.organization;
    if (organization?.name) {
      // @ts-ignore
      window.location = `${window.location.origin}/${organization.name}`;
    }
  }

  return (
    <Flex
      direction="column"
      align="center"
      gap={9}
      className={styles.container}
      {...props}
    >
      <Flex direction="column" align="center" gap={5}>
        <Headline size="t2">{title}</Headline>
        <Text size="regular" variant="secondary" className={styles.description}>
          {description}
        </Text>
      </Flex>
      <form onSubmit={handleSubmit(onSubmit)} className={styles.form}>
        <Flex direction="column" align="center" gap={9} className={styles.card}>
          <Field
            label={`${t.organization({ case: 'capital' })} name`}
            error={errors.title && String(errors.title?.message)}
          >
            <Input
              size="large"
              {...register('title')}
              placeholder={`Provide ${t.organization({ case: 'lower' })} name`}
            />
          </Field>
          <Field
            label={`${t.organization({ case: 'capital' })} URL`}
            error={errors.name && String(errors.name?.message)}
          >
            <Input
              size="large"
              {...register('name')}
              placeholder="raystack.org/"
            />
          </Field>
          <Button
            className={styles.submit}
            type="submit"
            data-test-id="frontier-sdk-create-workspace-btn"
            disabled={isSubmitting}
            loading={isSubmitting}
            loaderText="Creating..."
          >
            Create {t.organization({ case: 'lower' })}
          </Button>
        </Flex>
      </form>
    </Flex>
  );
};
