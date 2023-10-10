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

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import Skeleton from 'react-loading-skeleton';
import { toast } from 'sonner';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group, V1Beta1Role } from '~/src';
import { PERMISSIONS } from '~/utils';

const inviteSchema = yup.object({
  type: yup.string().required(),
  team: yup.string(),
  emails: yup.string().required()
});

type InviteSchemaType = yup.InferType<typeof inviteSchema>;

export const InviteMember = () => {
  const {
    watch,
    reset,
    control,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(inviteSchema)
  });
  const [teams, setTeams] = useState<V1Beta1Group[]>([]);
  const [roles, setRoles] = useState<V1Beta1Role[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate({ from: '/members/modal' });
  const { client, activeOrganization: organization } = useFrontier();

  const values = watch(['emails', 'team', 'type']);

  const isGroupRole = useMemo(() => {
    const role = values[2] && roles.find(r => r.id === values[2]);
    return role && role.scopes?.includes(PERMISSIONS.GroupNamespace);
  }, [roles, values]);

  const onSubmit = useCallback(
    async ({ emails, type, team }: InviteSchemaType) => {
      const emailList = emails
        .split(',')
        .map(e => e.trim())
        .filter(str => str.length > 0);

      if (!organization?.id) return;
      if (!emailList.length) return;
      if (!type) return;
      if (isGroupRole && !team) return;

      try {
        await client?.frontierServiceCreateOrganizationInvitation(
          organization?.id,
          {
            userIds: emailList,
            groupIds: isGroupRole && team ? [team] : undefined,
            roleIds: [type]
          }
        );
        toast.success('memebers added');

        navigate({ to: '/members' });
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    },
    [client, navigate, organization?.id, isGroupRole]
  );

  useEffect(() => {
    async function getInformation() {
      try {
        setIsLoading(true);

        if (!organization?.id) return;
        const {
          // @ts-ignore
          data: { roles: orgRoles }
        } = await client?.frontierServiceListOrganizationRoles(
          organization.id,
          {
            scopes: [
              PERMISSIONS.OrganizationNamespace,
              PERMISSIONS.GroupNamespace
            ]
          }
        );
        const {
          // @ts-ignore
          data: { roles }
        } = await client?.frontierServiceListRoles({
          scopes: [
            PERMISSIONS.OrganizationNamespace,
            PERMISSIONS.GroupNamespace
          ]
        });
        const {
          // @ts-ignore
          data: { groups }
        } = await client?.frontierServiceListOrganizationGroups(
          organization.id
        );
        setRoles([...roles, ...orgRoles]);
        setTeams(groups);
      } catch (err) {
        console.error(err);
      } finally {
        setIsLoading(false);
      }
    }
    getInformation();
  }, [client, organization?.id]);

  const isDisabled = useMemo(() => {
    const [emails, team, type] = values;
    const emailList =
      emails
        ?.split(',')
        .map((e: string) => e.trim())
        .filter(str => str.length > 0) || [];
    return (
      emailList.length <= 0 || !type || isSubmitting || (isGroupRole && !team)
    );
  }, [isGroupRole, isSubmitting, values]);

  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content style={{ padding: 0, maxWidth: '600px', width: '100%' }}>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex justify="between" style={{ padding: '16px 24px' }}>
            <Text size={6} style={{ fontWeight: '500' }}>
              Invite people
            </Text>

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              // @ts-ignore
              src={cross}
              onClick={() => navigate({ to: '/members' })}
            />
          </Flex>
          <Separator />
          <Flex
            direction="column"
            gap="medium"
            style={{ padding: '24px 32px' }}
          >
            <InputField label="Email">
              <Controller
                render={({ field }) => (
                  <textarea
                    {...field}
                    // @ts-ignore
                    style={{
                      appearance: 'none',
                      boxSizing: 'border-box',
                      margin: 0,
                      outline: 'none',
                      padding: 'var(--pd-8)',
                      height: 'auto',
                      width: '100%',

                      backgroundColor: 'var(--background-base)',
                      border: '0.5px solid var(--border-base)',
                      boxShadow: 'var(--shadow-xs)',
                      borderRadius: 'var(--br-4)',
                      color: 'var(--foreground-base)'
                    }}
                    placeholder="Enter comma seprated emails like abc@domain.com, bcd@domain.com"
                  />
                )}
                control={control}
                name="emails"
              />

              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.emails && String(errors.emails?.message)}
              </Text>
            </InputField>
            <InputField label="Invite as">
              {isLoading ? (
                <Skeleton height={'25px'} />
              ) : (
                <Controller
                  render={({ field }) => (
                    <Select {...field} onValueChange={field.onChange}>
                      <Select.Trigger className="w-[180px]">
                        <Select.Value placeholder="Select a role" />
                      </Select.Trigger>
                      <Select.Content
                        style={{ width: '100% !important', minWidth: '180px' }}
                      >
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
                  name="type"
                />
              )}
              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.type && String(errors.type?.message)}
              </Text>
            </InputField>

            <InputField label="Add to team">
              {isLoading ? (
                <Skeleton height={'25px'} />
              ) : (
                <Controller
                  rules={{ required: isGroupRole }}
                  render={({ field }) => (
                    <Select
                      {...field}
                      onValueChange={field.onChange}
                      disabled={!isGroupRole}
                    >
                      <Select.Trigger className="w-[180px]">
                        <Select.Value placeholder="Select a team" />
                      </Select.Trigger>
                      <Select.Content
                        style={{ width: '100% !important', minWidth: '180px' }}
                      >
                        <Select.Group>
                          {teams.map(t => (
                            <Select.Item value={t.id} key={t.id}>
                              {t.title}
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
            <Separator />
            <Flex justify="end">
              <Button
                variant="primary"
                size="medium"
                type="submit"
                disabled={isDisabled}
              >
                {isSubmitting ? 'sending...' : 'Send invite'}
              </Button>
            </Flex>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
