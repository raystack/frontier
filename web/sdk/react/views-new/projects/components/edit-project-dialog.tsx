'use client';

import { useEffect } from 'react';
import {
  Button,
  Flex,
  Text,
  Dialog,
  InputField,
  Radio
} from '@raystack/apsara-v1';
import { toastManager } from '@raystack/apsara-v1';
import * as yup from 'yup';
import { yupResolver } from '@hookform/resolvers/yup';
import { useForm } from 'react-hook-form';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  UpdateProjectRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import styles from './edit-project-dialog.module.css';
import { handleConnectError } from '~/utils/error';

const editProjectSchema = yup
  .object({
    title: yup.string().required('Project name is required'),
    privacy: yup.string().oneOf(['private', 'public']).required()
  })
  .required();

type FormData = yup.InferType<typeof editProjectSchema>;

export interface EditProjectPayload {
  projectId: string;
  title: string;
}

type DialogHandle = ReturnType<typeof Dialog.createHandle<EditProjectPayload>>;

export interface EditProjectDialogProps {
  handle: DialogHandle;
  refetch: () => void;
}

export function EditProjectDialog({ handle, refetch }: EditProjectDialogProps) {
  return (
    <Dialog handle={handle}>
      {({ payload }) => {
        const p = payload as EditProjectPayload | undefined;
        return (
          <Dialog.Content width={400}>
            {p ? (
              <EditProjectForm
                payload={p}
                handle={handle}
                refetch={refetch}
              />
            ) : null}
          </Dialog.Content>
        );
      }}
    </Dialog>
  );
}

interface EditProjectFormProps {
  payload: EditProjectPayload;
  handle: DialogHandle;
  refetch: () => void;
}

function EditProjectForm({ payload, handle, refetch }: EditProjectFormProps) {
  const {
    reset,
    handleSubmit,
    setValue,
    watch,
    formState: { errors, isSubmitting, isDirty },
    register
  } = useForm({
    resolver: yupResolver(editProjectSchema),
    defaultValues: {
      title: payload.title,
      privacy: 'private' as const
    }
  });

  const { activeOrganization: organization } = useFrontier();
  const privacy = watch('privacy');

  const { mutateAsync: updateProject } = useMutation(
    FrontierServiceQueries.updateProject
  );

  useEffect(() => {
    reset({
      title: payload.title,
      privacy: 'private'
    });
  }, [payload.projectId, payload.title, reset]);

  async function onSubmit(data: FormData) {
    if (!organization?.id || !payload.projectId) return;

    try {
      await updateProject(
        create(UpdateProjectRequestSchema, {
          id: payload.projectId,
          body: {
            title: data.title,
            orgId: organization.id
          }
        })
      );
      toastManager.add({ title: 'Project updated', type: 'success' });
      refetch();
      handle.close();
    } catch (error) {
      handleConnectError(error, {
        AlreadyExists: () => toastManager.add({ title: 'Project already exists', type: 'error' }),
        InvalidArgument: (err) => toastManager.add({ title: 'Invalid input', description: err.message, type: 'error' }),
        PermissionDenied: () => toastManager.add({ title: "You don't have permission to perform this action", type: 'error' }),
        NotFound: (err) => toastManager.add({ title: 'Not found', description: err.message, type: 'error' }),
        Default: (err) => toastManager.add({ title: 'Something went wrong', description: err.message, type: 'error' }),
      });
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Dialog.Header>
        <Dialog.Title>Edit project details</Dialog.Title>
      </Dialog.Header>
      <Dialog.Body>
        <Flex direction="column" gap={5}>
          <InputField
            label="Project name"
            size="large"
            error={errors.title && String(errors.title?.message)}
            {...register('title')}
            placeholder="Enter project name"
          />
          <Flex direction="column" gap={4}>
            <Text size="mini" weight="medium" variant="secondary">
              Project privacy
            </Text>
            <Radio.Group
              value={privacy}
              onValueChange={(val: string) =>
                setValue('privacy', val as 'private' | 'public', {
                  shouldDirty: true
                })
              }
            >
              <div className={styles.radioOptions}>
                <Flex gap={3}>
                  <Radio value="private" />
                  <Text size="small" variant="secondary">
                    Private
                  </Text>
                </Flex>
                <Flex gap={3}>
                  <Radio value="public" />
                  <Text size="small" variant="secondary">
                    Public
                  </Text>
                </Flex>
              </div>
            </Radio.Group>
          </Flex>
        </Flex>
      </Dialog.Body>
      <Dialog.Footer>
        <Flex gap={5} justify="end">
          <Button
            variant="outline"
            color="neutral"
            type="button"
            onClick={() => handle.close()}
            data-test-id="frontier-sdk-cancel-edit-project-btn"
          >
            Cancel
          </Button>
          <Button
            variant="solid"
            color="accent"
            type="submit"
            disabled={!isDirty}
            loading={isSubmitting}
            loaderText="Saving..."
            data-test-id="frontier-sdk-save-project-btn"
          >
            Save
          </Button>
        </Flex>
      </Dialog.Footer>
    </form>
  );
}
