import {
  Dialog,
  Flex,
  Image,
  InputField,
  Separator,
  Text,
  TextField
} from '@raystack/apsara';
import { Button } from '@raystack/apsara/v1';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate } from '@tanstack/react-router';
import { useEffect } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { toast } from '@raystack/apsara/v1';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import styles from '../organization.module.css';

const projectSchema = yup
  .object({
    title: yup.string().required(),
    name: yup
      .string()
      .required('name is a required field')
      .min(3, 'name is not valid, Min 3 characters allowed')
      .max(50, 'name is not valid, Max 50 characters allowed')
      .matches(
        /^[a-zA-Z0-9_-]{3,50}$/,
        "Only numbers, letters, '-', and '_' are allowed. Spaces are not allowed."
      ),
    org_id: yup.string().required()
  })
  .required();

type FormData = yup.InferType<typeof projectSchema>;

export const AddProject = () => {
  const {
    reset,
    control,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(projectSchema)
  });
  const navigate = useNavigate({ from: '/projects/modal' });
  const { client, activeOrganization: organization } = useFrontier();

  useEffect(() => {
    reset({ org_id: organization?.id });
  }, [organization, reset]);

  async function onSubmit(data: FormData) {
    if (!client) return;

    try {
      await client.frontierServiceCreateProject(data);
      toast.success('Project added');
      navigate({ to: '/projects' });
    } catch (err: unknown) {
      if (err instanceof Response && err?.status === 409) {
        setError('name', {
          message: 'Project name already exist, please enter unique name'
        });
      } else {
        toast.error('Something went wrong', {
          description: (err as Error)?.message
        });
      }
    }
  }

  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%', zIndex: '60' }}
        overlayClassname={styles.overlay}
      >
        <Flex justify="between" style={{ padding: '16px 24px' }}>
          <Text size={6} style={{ fontWeight: '500' }}>
            Add Project
          </Text>
          <Image
            alt="cross"
            // @ts-ignore
            src={cross}
            onClick={() => navigate({ to: '/projects' })}
            data-test-id="frontier-sdk-new-project-close-btn"
            style={{ cursor: 'pointer' }}
          />
        </Flex>
        <Separator />
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex
            direction="column"
            gap="medium"
            style={{ padding: '24px 32px' }}
          >
            <TextField
              name="orgId"
              defaultValue={organization?.id}
              hidden={true}
            />
            <InputField label="Project title">
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    // @ts-ignore
                    size="medium"
                    placeholder="Provide project title"
                  />
                )}
                control={control}
                name="title"
              />

              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.title && String(errors.title?.message)}
              </Text>
            </InputField>
            <InputField label="Project name">
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    // @ts-ignore
                    size="medium"
                    placeholder="Provide project name"
                  />
                )}
                control={control}
                name="name"
              />

              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.name && String(errors.name?.message)}
              </Text>
            </InputField>
          </Flex>
          <Separator />
          <Flex align="end" style={{ padding: 'var(--pd-16)' }}>
            <Button
              type="submit"
              data-test-id="frontier-sdk-add-project-btn"
              loading={isSubmitting}
              loaderText="Adding..."
            >
              Add project
            </Button>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
