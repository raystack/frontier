import {
  Flex,
  InputField,
  Text,
  TextField,
} from '@raystack/apsara';
import { Separator, Button, toast, Tooltip, Skeleton } from '@raystack/apsara/v1';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useEffect, useMemo } from 'react';
import { Controller, useForm } from 'react-hook-form';
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
    control,
    handleSubmit,
    formState: { errors, isSubmitting }
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
    <Flex direction="column" gap="large" style={{ paddingTop: '32px' }}>
      <form onSubmit={handleSubmit(onSubmit)}>
        <Flex direction="column" gap="medium" style={{ maxWidth: '320px' }}>
          {isLoading ? (
            <div>
              <Skeleton height={'16px'} />
              <Skeleton height={'32px'} />
            </div>
          ) : (
            <InputField label="Team title">
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    // @ts-ignore
                    size="medium"
                    placeholder="Provide team title"
                  />
                )}
                control={control}
                name="title"
              />

              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.title && String(errors.title?.message)}
              </Text>
            </InputField>
          )}
          {isLoading ? (
            <div>
              <Skeleton height={'16px'} />
              <Skeleton height={'32px'} />
            </div>
          ) : (
            <InputField label="Team name">
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    // @ts-ignore
                    size="medium"
                    disabled
                    placeholder="Provide team name"
                  />
                )}
                control={control}
                name="name"
              />
              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.name && String(errors.name?.message)}
              </Text>
            </InputField>
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
      <Separator />
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
    <Flex direction="column" gap="medium">
      {isLoading ? (
        <Skeleton height={'16px'} width={'50%'} />
      ) : (
        <Text size={3} style={{ color: 'var(--foreground-muted)' }}>
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
