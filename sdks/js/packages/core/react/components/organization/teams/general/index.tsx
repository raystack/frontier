import {
  Separator,
  Button,
  toast,
  Tooltip,
  Skeleton,
  Text,
  Flex,
  InputField
} from '@raystack/apsara/v1';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useEffect, useMemo } from 'react';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { usePermissions } from '~/react/hooks/usePermissions';
import { V1Beta1Group, V1Beta1Organization } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { AuthTooltipMessage } from '~/react/utils';

const teamSchema = yup
  .object({
    title: yup.string().required(),
    name: yup.string().required()
  })
  .required();

type FormData = yup.InferType<typeof teamSchema>;

interface GeneralTeamProps {
  team?: V1Beta1Group;
  organization?: V1Beta1Organization;
  isLoading?: boolean;
}

export const General = ({
  organization,
  team,
  isLoading: isTeamLoading
}: GeneralTeamProps) => {
  const {
    reset,
    handleSubmit,
    formState: { errors, isSubmitting },
    register
  } = useForm({
    resolver: yupResolver(teamSchema)
  });

  let { teamId } = useParams({ from: '/teams/$teamId' });
  const { client } = useFrontier();

  useEffect(() => {
    reset(team);
  }, [reset, team]);

  const resource = `app/group:${teamId}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.UpdatePermission,
        resource
      },
      {
        permission: PERMISSIONS.DeletePermission,
        resource
      }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!teamId
  );

  const { canUpdateGroup, canDeleteGroup } = useMemo(() => {
    return {
      canUpdateGroup: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      ),
      canDeleteGroup: shouldShowComponent(
        permissions,
        `${PERMISSIONS.DeletePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const isLoading = isTeamLoading || isPermissionsFetching;

  async function onSubmit(data: FormData) {
    if (!client) return;
    if (!organization?.id) return;
    if (!teamId) return;

    try {
      await client.frontierServiceUpdateGroup(organization?.id, teamId, data);
      toast.success('Team updated');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return (
    <Flex direction="column" gap={9} style={{ paddingTop: '32px' }}>
      <form onSubmit={handleSubmit(onSubmit)}>
        <Flex direction="column" gap={5} style={{ maxWidth: '320px' }}>
          {isLoading ? (
            <div>
              <Skeleton height={'16px'} />
              <Skeleton height={'32px'} />
            </div>
          ) : (
            <InputField
              label="Team title"
              size="large"
              error={errors.title && String(errors.title?.message)}
              {...register('title')}
              placeholder="Provide team title"
            />
          )}
          {isLoading ? (
            <div>
              <Skeleton height={'16px'} />
              <Skeleton height={'32px'} />
            </div>
          ) : (
            <InputField
              label="Team name"
              size="large"
              error={errors.name && String(errors.name?.message)}
              {...register('name')}
              disabled
              placeholder="Provide team name"
            />
          )}

          {isLoading ? (
            <Skeleton height={'32px'} width={'64px'} />
          ) : (
            <Tooltip message={AuthTooltipMessage} disabled={canUpdateGroup}>
              <Button
                type="submit"
                disabled={!canUpdateGroup}
                data-test-id="frontier-sdk-update-team-btn"
                loading={isSubmitting}
                loaderText="Updating..."
              >
                Update team
              </Button>
            </Tooltip>
          )}
        </Flex>
      </form>
      <Separator />
      <GeneralDeleteTeam
        organization={organization}
        canDeleteGroup={canDeleteGroup}
        isLoading={isLoading}
      />
    </Flex>
  );
};

interface GeneralDeleteTeamProps extends GeneralTeamProps {
  canDeleteGroup?: boolean;
}

export const GeneralDeleteTeam = ({
  canDeleteGroup,
  isLoading
}: GeneralDeleteTeamProps) => {
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const navigate = useNavigate({ from: '/teams/$teamId' });

  return (
    <Flex direction="column" gap={5}>
      {isLoading ? (
        <Skeleton height={'16px'} width={'50%'} />
      ) : (
        <Text size={3} variant="secondary">
          If you want to permanently delete this team and all of its data.
        </Text>
      )}
      {isLoading ? (
        <Skeleton height={'32px'} width={'64px'} />
      ) : (
        <Tooltip message={AuthTooltipMessage} disabled={canDeleteGroup}>
          <Button
            variant="solid"
            color="danger"
            type="submit"
            disabled={!canDeleteGroup}
            onClick={() =>
              navigate({
                to: `/teams/$teamId/delete`,
                params: { teamId: teamId }
              })
            }
            data-test-id="frontier-sdk-delete-team-btn"
          >
            Delete team
          </Button>
        </Tooltip>
      )}
    </Flex>
  );
};
