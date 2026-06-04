'use client';

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
import { useFrontier } from '../../../contexts/FrontierContext';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  CreateGroupRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';

const teamSchema = yup
  .object({
    title: yup.string().required('Team title is required'),
    name: yup
      .string()
      .required('Team name is required')
      .min(3, 'Name must be at least 3 characters')
      .max(50, 'Name must be at most 50 characters')
      .matches(
        /^[a-zA-Z0-9_-]{3,50}$/,
        "Only numbers, letters, '-', and '_' are allowed. Spaces are not allowed."
      )
  })
  .required();

type FormData = yup.InferType<typeof teamSchema>;

type DialogHandle = ReturnType<typeof Dialog.createHandle>;

export interface AddTeamDialogProps {
  handle: DialogHandle;
  refetch: () => void;
}

export function AddTeamDialog({ handle, refetch }: AddTeamDialogProps) {
  const {
    reset,
    handleSubmit,
    formState: { errors, isSubmitting },
    register
  } = useForm({
    resolver: yupResolver(teamSchema)
  });
  const { activeOrganization: organization } = useFrontier();

  const { mutateAsync: createTeam } = useMutation(
    FrontierServiceQueries.createGroup
  );

  const handleOpenChange = (open: boolean) => {
    if (!open) reset();
  };

  async function onSubmit(data: FormData) {
    if (!organization?.id) return;

    try {
      await createTeam(
        create(CreateGroupRequestSchema, {
          orgId: organization.id,
          body: {
            title: data.title,
            name: data.name
          }
        })
      );
      toastManager.add({ title: 'Team added', type: 'success' });
      refetch();
      handle.close();
    } catch (error) {
      toastManager.add({
        title: 'Something went wrong',
        description:
          error instanceof Error ? error.message : 'Failed to create team',
        type: 'error'
      });
    }
  }

  return (
    <Dialog handle={handle} onOpenChange={handleOpenChange}>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title>Add Team</Dialog.Title>
        </Dialog.Header>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Body>
            <Flex direction="column" gap={5}>
              <Field
                label="Team title"
                error={errors.title && String(errors.title?.message)}
              >
                <Input
                  size="large"
                  {...register('title')}
                  placeholder="Provide team title"
                />
              </Field>
              <Field
                label="Team name"
                error={errors.name && String(errors.name?.message)}
              >
                <Input
                  size="large"
                  {...register('name')}
                  placeholder="Provide team name"
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
                data-test-id="frontier-sdk-cancel-add-team-btn"
              >
                Cancel
              </Button>
              <Button
                variant="solid"
                color="accent"
                type="submit"
                loading={isSubmitting}
                loaderText="Adding..."
                data-test-id="frontier-sdk-add-team-btn"
              >
                Add team
              </Button>
            </Flex>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
}
