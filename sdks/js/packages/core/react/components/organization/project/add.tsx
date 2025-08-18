import { useEffect } from 'react';
import {
  Button,
  toast,
  Image,
  Text,
  Flex,
  Dialog,
  InputField
} from '@raystack/apsara';
import * as yup from 'yup';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate } from '@tanstack/react-router';
import { useForm } from 'react-hook-form';
import { useFrontier } from '~/react/contexts/FrontierContext';
import cross from '~/react/assets/cross.svg';
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
    formState: { errors, isSubmitting },
    register
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
          message: 'Project name already exists. Please enter a unique name.'
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
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%' }}
        overlayClassName={styles.overlay}
      >
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="large" style={{ fontWeight: '500' }}>
              Add Project
            </Text>
            <Image
              alt="cross"
              src={cross as unknown as string}
              onClick={() => navigate({ to: '/projects' })}
              data-test-id="frontier-sdk-new-project-close-btn"
              style={{ cursor: 'pointer' }}
            />
          </Flex>
        </Dialog.Header>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Body>
            <Flex direction="column" gap={5}>
              <div style={{ display: 'none' }}>
                <InputField name="orgId" defaultValue={organization?.id} />
              </div>
              <InputField
                label="Project title"
                size="large"
                error={errors.title && String(errors.title?.message)}
                {...register('title')}
                placeholder="Provide project title"
              />
              <InputField
                label="Project name"
                size="large"
                error={errors.name && String(errors.name?.message)}
                {...register('name')}
                placeholder="Provide project name"
              />
            </Flex>
          </Dialog.Body>
          <Dialog.Footer>
            <Flex align="end">
              <Button
                type="submit"
                data-test-id="frontier-sdk-add-project-btn"
                loading={isSubmitting}
                loaderText="Adding..."
              >
                Add project
              </Button>
            </Flex>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
