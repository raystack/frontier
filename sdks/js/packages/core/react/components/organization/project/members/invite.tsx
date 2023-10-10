import { yupResolver } from '@hookform/resolvers/yup';
import {
  Button,
  Dialog,
  Flex,
  Image,
  InputField,
  Select,
  Separator,
  Text
} from '@raystack/apsara';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useCallback, useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import Skeleton from 'react-loading-skeleton';
import { toast } from 'sonner';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group, V1Beta1PolicyRequestBody, V1Beta1Role } from '~/src';
import { PERMISSIONS } from '~/utils';

const inviteSchema = yup.object({
  team: yup.string().required(),
  role: yup.string().required()
});

type InviteSchemaType = yup.InferType<typeof inviteSchema>;

export const InviteProjectTeam = () => {
  let { projectId } = useParams({ from: '/projects/$projectId/invite' });
  const navigate = useNavigate({ from: '/projects/$projectId/invite' });
  const [roles, setRoles] = useState<V1Beta1Role[]>([]);
  const [teams, setTeams] = useState<V1Beta1Group[]>([]);

  const [isRolesLoading, setIsRolesLoading] = useState(false);
  const [isTeamsLoading, setIsTeamsLoading] = useState(false);
  const { client, activeOrganization: organization } = useFrontier();

  const {
    watch,
    reset,
    control,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(inviteSchema)
  });

  const getTeams = useCallback(async () => {
    try {
      setIsTeamsLoading(true);
      if (!organization?.id) return;
      const {
        // @ts-ignore
        data: { groups }
      } = await client?.frontierServiceListOrganizationGroups(organization.id);

      setTeams(groups);
    } catch (err) {
      console.error(err);
    } finally {
      setIsTeamsLoading(false);
    }
  }, [client, organization?.id]);

  const getRoles = useCallback(async () => {
    try {
      setIsRolesLoading(true);
      if (!organization?.id) return;
      const {
        // @ts-ignore
        data: { roles: orgRoles }
      } = await client?.frontierServiceListOrganizationRoles(organization.id, {
        scopes: [PERMISSIONS.ProjectNamespace]
      });
      const {
        // @ts-ignore
        data: { roles }
      } = await client?.frontierServiceListRoles({
        scopes: [PERMISSIONS.ProjectNamespace]
      });
      setRoles([...roles, ...orgRoles]);
    } catch (err) {
      console.error(err);
    } finally {
      setIsRolesLoading(false);
    }
  }, [client, organization?.id]);

  useEffect(() => {
    getRoles();
    getTeams();
  }, [getRoles, getTeams, organization?.id]);

  async function onSubmit({ role, team }: InviteSchemaType) {
    if (!team || !role || !projectId) return;
    try {
      const resource = `${PERMISSIONS.ProjectNamespace}:${projectId}`;
      const principal = `${PERMISSIONS.GroupPrincipal}:${team}`;

      const policy: V1Beta1PolicyRequestBody = {
        roleId: role,
        resource,
        principal
      };
      await client?.frontierServiceCreatePolicy(policy);
      toast.success('team added');
      navigate({
        to: '/projects/$projectId',
        params: { projectId }
      });
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content style={{ padding: 0, maxWidth: '600px', width: '100%' }}>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex justify="between" style={{ padding: '16px 24px' }}>
            <Text size={6} style={{ fontWeight: '500' }}>
              Add Team
            </Text>

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              // @ts-ignore
              src={cross}
              onClick={() =>
                navigate({ to: '/projects/$projectId', params: { projectId } })
              }
            />
          </Flex>
          <Separator />
          <Flex
            direction="column"
            gap="medium"
            style={{ padding: '24px 32px' }}
          >
            <InputField label="Teams">
              {isTeamsLoading ? (
                <Skeleton height={'25px'} />
              ) : (
                <Controller
                  render={({ field }) => (
                    <Select {...field} onValueChange={field.onChange}>
                      <Select.Trigger className="w-[180px]">
                        <Select.Value placeholder="Select a team" />
                      </Select.Trigger>
                      <Select.Content
                        style={{ width: '100% !important', minWidth: '180px' }}
                      >
                        <Select.Group>
                          {!teams.length && (
                            <Select.Label>No teams available</Select.Label>
                          )}
                          {teams.map(team => (
                            <Select.Item value={team.id} key={team.id}>
                              {team.title || team.name}
                            </Select.Item>
                          ))}
                        </Select.Group>
                      </Select.Content>
                    </Select>
                  )}
                  control={control}
                  name="team"
                />
              )}
              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.team && String(errors.team?.message)}
              </Text>
            </InputField>
            <InputField label="Invite as">
              {isRolesLoading ? (
                <Skeleton height={'25px'} />
              ) : (
                <Controller
                  render={({ field }) => (
                    <Select {...field} onValueChange={field.onChange}>
                      <Select.Trigger className="w-[180px]">
                        <Select.Value placeholder="Select a role" />
                      </Select.Trigger>
                      <Select.Content style={{ width: '100% !important' }}>
                        <Select.Group>
                          {!roles.length && (
                            <Select.Label>No roles available</Select.Label>
                          )}
                          {roles.map(role => (
                            <Select.Item value={role.id} key={role.id}>
                              {role.title || role.name}
                            </Select.Item>
                          ))}
                        </Select.Group>
                      </Select.Content>
                    </Select>
                  )}
                  control={control}
                  name="role"
                />
              )}
              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.role && String(errors.role?.message)}
              </Text>
            </InputField>
            <Separator />
            <Flex justify="end">
              <Button variant="primary" size="medium" type="submit">
                {isSubmitting ? 'adding...' : 'Add Team'}
              </Button>
            </Flex>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
