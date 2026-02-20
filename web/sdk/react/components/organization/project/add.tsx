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
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  CreateProjectRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import cross from '~/react/assets/cross.svg';
import styles from '../organization.module.css';
import slugify from 'slugify';
import { generateHashFromString } from '~/react/utils';
import { ConnectError, Code } from '@connectrpc/connect';

const projectSchema = yup
  .object({
    title: yup.string().required(),
    org_id: yup.string().required()
  })
  .required();

type FormData = yup.InferType<typeof projectSchema>;

export const AddProject = () => {
  const {
    reset,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
    register
  } = useForm({
    resolver: yupResolver(projectSchema)
  });
  const navigate = useNavigate({ from: '/projects/modal' });
  const { activeOrganization: organization } = useFrontier();

  useEffect(() => {
    reset({ org_id: organization?.id });
  }, [organization, reset]);

  const { mutateAsync: createProject } = useMutation(
    FrontierServiceQueries.createProject,
    {
      onSuccess: () => {
        toast.success('Project added');
        navigate({ to: '/projects' });
      }
    }
  );

  async function onSubmit(data: FormData) {
    if (!organization?.id) return;
    const slug = slugify(data.title, { lower: true, strict: true });
    const suffix = generateHashFromString(organization.id);
    const name = `${slug}-${suffix}`;
    try {
      await createProject(
        create(CreateProjectRequestSchema, {
          body: {
            title: data.title,
            name,
            orgId: organization.id
          }
        })
      );
    } catch (error) {
      if (error instanceof ConnectError && error.code === Code.AlreadyExists) {
        setError('title', {
          message:
            'A project with a similar title already exist. Please tweak the title and try again.'
        });
      } else {
        toast.error('Something went wrong', {
          description:
            error instanceof Error ? error.message : 'Failed to create project'
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
