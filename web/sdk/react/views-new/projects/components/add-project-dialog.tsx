'use client';

import {
  Button,
  Flex,
  Dialog,
  InputField
} from '@raystack/apsara-v1';
import { toastManager } from '@raystack/apsara-v1';
import * as yup from 'yup';
import { yupResolver } from '@hookform/resolvers/yup';
import { useForm } from 'react-hook-form';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  CreateProjectRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import slugify from 'slugify';
import { generateHashFromString } from '../../../utils';
import { handleConnectError } from '~/utils/error';

const projectSchema = yup
  .object({
    title: yup.string().required('Project title is required')
  })
  .required();

type FormData = yup.InferType<typeof projectSchema>;

type DialogHandle = ReturnType<typeof Dialog.createHandle>;

export interface AddProjectDialogProps {
  handle: DialogHandle;
  refetch: () => void;
}

export function AddProjectDialog({ handle, refetch }: AddProjectDialogProps) {
  const {
    reset,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
    register
  } = useForm({
    resolver: yupResolver(projectSchema)
  });
  const { activeOrganization: organization } = useFrontier();

  const { mutateAsync: createProject } = useMutation(
    FrontierServiceQueries.createProject
  );

  const handleOpenChange = (open: boolean) => {
    if (!open) reset();
  };

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
      toastManager.add({ title: 'Project added', type: 'success' });
      refetch();
      handle.close();
    } catch (error) {
      handleConnectError(error, {
        AlreadyExists: () => setError('title', {
          message: 'A project with a similar title already exists. Please tweak the title and try again.'
        }),
        InvalidArgument: (err) => toastManager.add({ title: 'Invalid input', description: err.message, type: 'error' }),
        PermissionDenied: () => toastManager.add({ title: "You don't have permission to perform this action", type: 'error' }),
        Default: (err) => toastManager.add({ title: 'Something went wrong', description: err.message, type: 'error' }),
      });
    }
  }

  return (
    <Dialog handle={handle} onOpenChange={handleOpenChange}>
      <Dialog.Content width={400}>
        <Dialog.Header>
          <Dialog.Title>Add Project</Dialog.Title>
        </Dialog.Header>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Body>
            <Flex direction="column" gap={5}>
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
            <Flex gap={5} justify="end">
              <Button
                variant="outline"
                color="neutral"
                type="button"
                onClick={() => handle.close()}
                data-test-id="frontier-sdk-cancel-add-project-btn"
              >
                Cancel
              </Button>
              <Button
                variant="solid"
                color="accent"
                type="submit"
                loading={isSubmitting}
                loaderText="Adding..."
                data-test-id="frontier-sdk-add-project-btn"
              >
                Add project
              </Button>
            </Flex>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
}
