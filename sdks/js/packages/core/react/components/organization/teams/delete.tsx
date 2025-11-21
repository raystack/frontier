import {
  Button,
  Checkbox,
  Skeleton,
  Image,
  Text,
  Flex,
  Dialog,
  toast,
  InputField
} from '@raystack/apsara';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries, DeleteGroupRequestSchema, GetGroupRequestSchema } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import styles from '../organization.module.css';

const teamSchema = yup
  .object({
    title: yup.string()
  })
  .required();

export const DeleteTeam = () => {
  const {
    watch,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
    register
  } = useForm({
    resolver: yupResolver(teamSchema)
  });
  let { teamId } = useParams({ from: `/teams/$teamId/delete` });
  const navigate = useNavigate();
  const [isAcknowledged, setIsAcknowledged] = useState(false);

  const { activeOrganization: organization } = useFrontier();

  // Get team details using Connect RPC
  const { data: teamData, isLoading: isTeamLoading, error: teamError } = useQuery(
    FrontierServiceQueries.getGroup,
    create(GetGroupRequestSchema, { id: teamId || '', orgId: organization?.id || '' }),
    { enabled: !!organization?.id && !!teamId }
  );

  const team = teamData?.group;

  // Handle team error
  useEffect(() => {
    if (teamError) {
      toast.error('Something went wrong', {
        description: teamError.message
      });
    }
  }, [teamError]);

  // Delete team using Connect RPC
  const deleteTeamMutation = useMutation(FrontierServiceQueries.deleteGroup, {
    onSuccess: () => {
      toast.success('team deleted');
      navigate({ to: '/teams' });
    },
    onError: (error) => {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  });

  function onSubmit(data: { title?: string }) {
    if (!organization?.id) return;
    if (!teamId) return;

    if (data.title !== team?.title)
      return setError('title', { message: 'Team title does not match' });

    const request = create(DeleteGroupRequestSchema, {
      id: teamId,
      orgId: organization.id
    });

    deleteTeamMutation.mutate(request);
  }

  const title = watch('title', '');
  return (
    <Dialog open={true}>
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%' }}
        overlayClassName={styles.overlay}
      >
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="large" weight="medium">
              Verify team deletion
            </Text>
            <Image
              alt="cross"
              src={cross as unknown as string}
              onClick={() =>
                navigate({
                  to: `/teams/$teamId`,
                  params: {
                    teamId
                  }
                })
              }
              style={{ cursor: 'pointer' }}
              data-test-id="frontier-sdk-delete-team-close-btn"
            />
          </Flex>
        </Dialog.Header>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Body>
            <Flex direction="column" gap={5}>
              {isTeamLoading ? (
                <>
                  <Skeleton height={'16px'} />
                  <Skeleton width={'50%'} height={'16px'} />
                  <Skeleton height={'32px'} />
                  <Skeleton height={'16px'} />
                  <Skeleton height={'32px'} />
                </>
              ) : (
                <>
                  <Text size={2}>
                    This action can not be undone. This will permanently delete
                    team <b>{team?.title}</b>.
                  </Text>

                  <InputField
                    label="Please enter the title of the team to confirm."
                    size="large"
                    error={errors.title && String(errors.title?.message)}
                    {...register('title')}
                    placeholder="Enter the team title"
                  />

                  <Flex gap="small">
                    <Checkbox
                      checked={isAcknowledged}
                      onCheckedChange={v => setIsAcknowledged(v === true)}
                      data-test-id="frontier-sdk-delete-team-checkbox"
                    />
                    <Text size={2}>
                      I acknowledge and understand that all of the team data will be deleted
                      and want to proceed.
                    </Text>
                  </Flex>
                  <Button
                    variant="solid"
                    color="danger"
                    disabled={!title || !isAcknowledged}
                    type="submit"
                    style={{ width: '100%' }}
                    data-test-id="frontier-sdk-delete-team-btn-general"
                    loading={deleteTeamMutation.isPending || isSubmitting}
                    loaderText="Deleting..."
                  >
                    Delete this team
                  </Button>
                </>
              )}
            </Flex>
          </Dialog.Body>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
