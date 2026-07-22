'use client';

import { useEffect } from 'react';
import {
  Button,
  Flex,
  Dialog,
  Field,
  Input
} from '@raystack/apsara';
import { toastManager } from '@raystack/apsara';
import * as yup from 'yup';
import { yupResolver } from '@hookform/resolvers/yup';
import { useForm } from 'react-hook-form';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  UpdateProjectRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { handleConnectError } from '~/utils/error';

const editProjectSchema = yup
  .object({
    title: yup.string().required('Project name is required')
  })
  .required();

type FormData = yup.InferType<typeof editProjectSchema>;

export interface EditProjectPayload {
  projectId: string;
  name: string;
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
          <Dialog.Content>
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
    formState: { errors, isSubmitting, isDirty },
    register
  } = useForm({
    resolver: yupResolver(editProjectSchema),
    defaultValues: {
      title: payload.title
    }
  });

  const { mutateAsync: updateProject } = useMutation(
    FrontierServiceQueries.updateProject
  );

  useEffect(() => {
    reset({
      title: payload.title
    });
  }, [payload.projectId, payload.title, reset]);

  async function onSubmit(data: FormData) {
    if (!payload.projectId) return;

    try {
      await updateProject(
        create(UpdateProjectRequestSchema, {
          id: payload.projectId,
          body: {
            name: payload.name,
            title: data.title
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
    <form noValidate onSubmit={handleSubmit(onSubmit)}>
      <Dialog.Header>
        <Dialog.Title>Edit project details</Dialog.Title>
      </Dialog.Header>
      <Dialog.Body>
        <Flex direction="column" gap={5}>
          <Field
            label="Project name"
            error={errors.title && String(errors.title?.message)}
          >
            <Input
              size="large"
              {...register('title')}
              placeholder="Enter project name"
            />
          </Field>
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
