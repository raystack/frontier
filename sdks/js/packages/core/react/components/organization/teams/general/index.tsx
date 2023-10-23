import {
  Button,
  Flex,
  InputField,
  Separator,
  Text,
  TextField
} from '@raystack/apsara';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { usePermissions } from '~/react/hooks/usePermissions';
import { V1Beta1Group, V1Beta1Organization } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import Skeleton from 'react-loading-skeleton';

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
  const listOfPermissionsToCheck = [
    {
      permission: PERMISSIONS.UpdatePermission,
      resource
    },
    {
      permission: PERMISSIONS.DeletePermission,
      resource
    }
  ];

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
          <InputField label="Team title">
            {isLoading ? (
              <Skeleton height={'32px'} />
            ) : (
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
            )}

            <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
              {errors.title && String(errors.title?.message)}
            </Text>
          </InputField>
          <InputField label="Team name">
            {isLoading ? (
              <Skeleton height={'32px'} />
            ) : (
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
            )}

            <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
              {errors.name && String(errors.name?.message)}
            </Text>
          </InputField>

          {canUpdateGroup ? (
            <Button variant="primary" size="medium" type="submit">
              {isSubmitting ? 'updating...' : 'Update team'}
            </Button>
          ) : null}
        </Flex>
      </form>
      <Separator />
      {canDeleteGroup ? (
        <>
          <GeneralDeleteTeam organization={organization} />
          <Separator />
        </>
      ) : null}
    </Flex>
  );
};

export const GeneralDeleteTeam = ({ organization }: GeneralTeamProps) => {
  const { client } = useFrontier();
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const navigate = useNavigate({ from: '/teams/$teamId' });
  const {
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm();

  const organizationId = organization?.id;

  const onDeleteOrganization = useCallback(async () => {
    if (!organizationId || !teamId) return;
    try {
      await client?.frontierServiceDeleteGroup(organizationId, teamId);
      navigate({ to: '/teams' });
    } catch ({ error }: any) {
      console.log(error);
      toast.error('Something went wrong', {
        description: `${error.message}`
      });
    }
  }, [client, navigate, organizationId, teamId]);

  return (
    <Flex direction="column" gap="medium">
      <Text size={3} style={{ color: 'var(--foreground-muted)' }}>
        If you want to permanently delete this team and all of its data.
      </Text>
      <Button
        variant="danger"
        type="submit"
        size="medium"
        onClick={() =>
          navigate({ to: `/teams/$teamId/delete`, params: { teamId: teamId } })
        }
      >
        {isSubmitting ? 'deleting...' : 'Delete team'}
      </Button>
    </Flex>
  );
};
