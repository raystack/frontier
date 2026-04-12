'use client';

import { useEffect } from 'react';
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
  UpdateGroupRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';

const editTeamSchema = yup
  .object({
    title: yup.string().required('Team title is required')
  })
  .required();

type FormData = yup.InferType<typeof editTeamSchema>;

export interface EditTeamPayload {
  teamId: string;
  title: string;
  name: string;
}

type DialogHandle = ReturnType<typeof Dialog.createHandle<EditTeamPayload>>;

export interface EditTeamDialogProps {
  handle: DialogHandle;
  refetch: () => void;
}

export function EditTeamDialog({ handle, refetch }: EditTeamDialogProps) {
  return (
    <Dialog handle={handle}>
      {({ payload }) => {
        const p = payload as EditTeamPayload | undefined;
        return (
          <Dialog.Content width={400}>
            {p ? (
              <EditTeamForm
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

interface EditTeamFormProps {
  payload: EditTeamPayload;
  handle: DialogHandle;
  refetch: () => void;
}

function EditTeamForm({ payload, handle, refetch }: EditTeamFormProps) {
  const {
    reset,
    handleSubmit,
    formState: { errors, isSubmitting, isDirty },
    register
  } = useForm({
    resolver: yupResolver(editTeamSchema),
    defaultValues: {
      title: payload.title
    }
  });

  const { activeOrganization: organization } = useFrontier();

  const { mutateAsync: updateTeam } = useMutation(
    FrontierServiceQueries.updateGroup
  );

  useEffect(() => {
    reset({ title: payload.title });
  }, [payload.teamId, payload.title, reset]);

  async function onSubmit(data: FormData) {
    if (!organization?.id || !payload.teamId) return;

    try {
      await updateTeam(
        create(UpdateGroupRequestSchema, {
          id: payload.teamId,
          orgId: organization.id,
          body: {
            title: data.title,
            name: payload.name
          }
        })
      );
      toastManager.add({ title: 'Team updated', type: 'success' });
      refetch();
      handle.close();
    } catch (error) {
      toastManager.add({
        title: 'Something went wrong',
        description:
          error instanceof Error ? error.message : 'Failed to update team',
        type: 'error'
      });
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Dialog.Header>
        <Dialog.Title>Edit team</Dialog.Title>
      </Dialog.Header>
      <Dialog.Body>
        <Flex direction="column" gap={5}>
          <InputField
            label="Team title"
            size="large"
            error={errors.title && String(errors.title?.message)}
            {...register('title')}
            placeholder="Provide team title"
          />
          <InputField
            label="Team name"
            size="large"
            value={payload.name}
            disabled
            placeholder="Team name"
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
            data-test-id="frontier-sdk-cancel-edit-team-btn"
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
            data-test-id="frontier-sdk-save-team-btn"
          >
            Save
          </Button>
        </Flex>
      </Dialog.Footer>
    </form>
  );
}
