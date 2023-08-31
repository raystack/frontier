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
import { useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group, V1Beta1Organization, V1Beta1Role } from '~/src';

const inviteSchema = yup.object({
  type: yup.string(),
  team: yup.string(),
  emails: yup.string().required()
});

export const InviteMember = () => {
  const {
    reset,
    control,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(inviteSchema)
  });
  const [teams, setTeams] = useState<V1Beta1Group[]>([]);
  const [roles, setRoles] = useState<V1Beta1Role[]>([]);
  const [selectedRole, setRole] = useState<string>();
  const [selectedTeam, setTeam] = useState<string>();
  const navigate = useNavigate({ from: '/members/modal' });
  const { client, activeOrganization: organization } = useFrontier();

  async function onSubmit({ emails }: any) {
    const emailList = emails.split(',').map((e: string) => e.trim());

    if (!organization?.id) return;
    if (!emailList.length) return;
    if (!selectedRole) return;
    if (!selectedTeam) return;

    try {
      await client?.frontierServiceCreateOrganizationInvitation(
        organization?.id,
        {
          userIds: emailList,
          groupIds: [selectedTeam],
          roleIds: [selectedRole]
        }
      );
      toast.success('memebers added');

      navigate({ to: '/members' });
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }
  useEffect(() => {
    async function getInformation() {
      if (!organization?.id) return;

      const {
        // @ts-ignore
        data: { roles: orgRoles }
      } = await client?.frontierServiceListOrganizationRoles(organization.id);
      const {
        // @ts-ignore
        data: { roles }
      } = await client?.frontierServiceListRoles();
      const {
        // @ts-ignore
        data: { groups }
      } = await client?.frontierServiceListOrganizationGroups(organization.id);
      setRoles([...roles, ...orgRoles]);
      setTeams(groups);
    }
    getInformation();
  }, [client, organization?.id]);

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
            <InputField label="Invite as">
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
              <Controller
                render={({ field }) => (
                  <Select
                    {...field}
                    onValueChange={(value: string) => setRole(value)}
                  >
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
                name="type"
              />

              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.emails && String(errors.emails?.message)}
              </Text>
            </InputField>

            <InputField label="Add to team">
              <Controller
                render={({ field }) => (
                  <Select
                    {...field}
                    onValueChange={(value: string) => setTeam(value)}
                  >
                    <Select.Trigger className="w-[180px]">
                      <Select.Value placeholder="Select a team" />
                    </Select.Trigger>
                    <Select.Content style={{ width: '100% !important' }}>
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

              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.emails && String(errors.emails?.message)}
              </Text>
            </InputField>
            <Separator />
            <Flex justify="end">
              <Button variant="primary" size="medium" type="submit">
                {isSubmitting ? 'sending...' : 'Send invite'}
              </Button>
            </Flex>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
