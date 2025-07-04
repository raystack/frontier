import { yupResolver } from '@hookform/resolvers/yup';
import {
  Button,
  toast,
  Skeleton,
  Image,
  Text,
  Flex,
  Dialog,
  Select,
  Label
} from '@raystack/apsara/v1';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1PolicyRequestBody, V1Beta1Role, V1Beta1User } from '~/src';
import { PERMISSIONS, filterUsersfromUsers } from '~/utils';
import cross from '~/react/assets/cross.svg';
import styles from '../../organization.module.css';
import { handleSelectValueChange } from '~/react/utils';

const inviteSchema = yup.object({
  userId: yup.string().required('Member is required'),
  role: yup.string().required('Role is required')
});

type InviteSchemaType = yup.InferType<typeof inviteSchema>;

export const InviteTeamMembers = () => {
  let { teamId } = useParams({ from: '/teams/$teamId/invite' });
  const navigate = useNavigate({ from: '/teams/$teamId/invite' });
  const [roles, setRoles] = useState<V1Beta1Role[]>([]);

  const [orgMembers, setOrgMembers] = useState<V1Beta1User[]>([]);
  const [isOrgMembersLoading, setIsOrgMembersLoading] = useState(false);

  const [members, setMembers] = useState<V1Beta1User[]>([]);

  const [isTeamMembersLoading, setIsTeamMembersLoading] = useState(false);

  const [isRolesLoading, setIsRolesLoading] = useState(false);
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

  useEffect(() => {
    async function getOrganizationMembers() {
      if (!organization?.id) return;
      try {
        setIsOrgMembersLoading(true);
        const {
          // @ts-ignore
          data: { users }
        } = await client?.frontierServiceListOrganizationUsers(
          organization?.id
        );
        setOrgMembers(users);
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      } finally {
        setIsOrgMembersLoading(false);
      }
    }
    getOrganizationMembers();
  }, [client, organization?.id]);

  useEffect(() => {
    async function getTeamMembers() {
      if (!organization?.id || !teamId) return;
      try {
        setIsTeamMembersLoading(true);
        const {
          // @ts-ignore
          data: { users, role_pairs = [] }
        } = await client?.frontierServiceListGroupUsers(
          organization?.id,
          teamId,
          { with_roles: true }
        );

        setMembers(users);
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      } finally {
        setIsTeamMembersLoading(false);
      }
    }
    getTeamMembers();
  }, [client, organization?.id, teamId]);

  const getRoles = useCallback(async () => {
    try {
      setIsRolesLoading(true);
      if (!organization?.id) return;
      const {
        // @ts-ignore
        data: { roles: orgRoles }
      } = await client?.frontierServiceListOrganizationRoles(organization.id, {
        scopes: [PERMISSIONS.GroupNamespace]
      });
      const {
        // @ts-ignore
        data: { roles }
      } = await client?.frontierServiceListRoles({
        scopes: [PERMISSIONS.GroupNamespace]
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
  }, [getRoles, organization?.id]);

  const addGroupTeamPolicy = useCallback(
    async (roleId: string, userId: string) => {
      const role = roles.find(r => r.id === roleId);
      if (role?.name && role.name !== PERMISSIONS.RoleGroupMember) {
        const resource = `${PERMISSIONS.GroupPrincipal}:${teamId}`;
        const principal = `${PERMISSIONS.UserPrincipal}:${userId}`;
        const policy: V1Beta1PolicyRequestBody = {
          role_id: roleId,
          resource,
          principal
        };
        await client?.frontierServiceCreatePolicy(policy);
      }
    },
    [client, roles, teamId]
  );

  async function onSubmit({ role, userId }: InviteSchemaType) {
    if (!userId || !role || !organization?.id) return;
    try {
      await client?.frontierServiceAddGroupUsers(organization?.id, teamId, {
        user_ids: [userId]
      });
      await addGroupTeamPolicy(role, userId);
      toast.success('member added');
      navigate({
        to: '/teams/$teamId',
        params: { teamId }
      });
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  const invitableUser = useMemo(
    () => filterUsersfromUsers(orgMembers, members) || [],
    [orgMembers, members]
  );

  const isUserLoading = isOrgMembersLoading || isTeamMembersLoading;

  return (
    <Dialog open={true}>
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%' }}
        overlayClassName={styles.overlay}
      >
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="large" weight="medium">
              Add Member
            </Text>

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              src={cross as unknown as string}
              onClick={() =>
                navigate({ to: '/teams/$teamId', params: { teamId } })
              }
              data-test-id="frontier-sdk-invite-team-members-close-btn"
            />
          </Flex>
        </Dialog.Header>
        <Dialog.Body>
          <form onSubmit={handleSubmit(onSubmit)}>
            <Flex direction="column" gap={5}>
              <Flex direction="column" gap={2}>
                <Label>Members</Label>
                {isUserLoading ? (
                  <Skeleton height={'25px'} />
                ) : (
                  <Controller
                    render={({ field: { onChange, ref, ...rest } }) => (
                      <Select
                        {...rest}
                        onValueChange={handleSelectValueChange(onChange)}
                      >
                        <Select.Trigger ref={ref}>
                          <Select.Value placeholder="Select members" />
                        </Select.Trigger>
                        <Select.Content style={{ width: '100% !important' }}>
                          <Select.Viewport style={{ maxHeight: '300px' }}>
                            <Select.Group>
                              {!invitableUser.length && (
                                <Text className={styles.noSelectItem}>
                                  No member to invite
                                </Text>
                              )}
                              {invitableUser.map(user => (
                                <Select.Item
                                  value={user.id || ''}
                                  key={user.id}
                                >
                                  {user.title || user.email}
                                </Select.Item>
                              ))}
                            </Select.Group>
                          </Select.Viewport>
                        </Select.Content>
                      </Select>
                    )}
                    control={control}
                    name="userId"
                  />
                )}
                <Text size="mini" variant="danger">
                  {errors.userId && String(errors.userId?.message)}
                </Text>
              </Flex>
              <Flex direction="column" gap={2}>
                <Label>Invite as</Label>
                {isRolesLoading ? (
                  <Skeleton height={'25px'} />
                ) : (
                  <Controller
                    render={({ field: { onChange, ref, ...rest } }) => (
                      <Select
                        {...rest}
                        onValueChange={handleSelectValueChange(onChange)}
                      >
                        <Select.Trigger ref={ref}>
                          <Select.Value placeholder="Select a role" />
                        </Select.Trigger>
                        <Select.Content style={{ width: '100% !important' }}>
                          <Select.Group>
                            {!roles.length && (
                              <Text className={styles.noSelectItem}>
                                No roles available
                              </Text>
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
                <Text size="mini" variant="danger">
                  {errors.role && String(errors.role?.message)}
                </Text>
              </Flex>
              <Flex justify="end">
                <Button
                  type="submit"
                  data-test-id="frontier-sdk-add-team-members-btn"
                  disabled={isUserLoading || isRolesLoading}
                  loading={isSubmitting}
                  loaderText="Adding..."
                >
                  Add Member
                </Button>
              </Flex>
            </Flex>
          </form>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
