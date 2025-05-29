import {
  Dialog,
  Flex,
  InputField,
  Select,
} from '@raystack/apsara';
import { Button, Separator, toast, Skeleton, Image, Text } from '@raystack/apsara/v1';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group, V1Beta1Role } from '~/src';
import { PERMISSIONS } from '~/utils';
import styles from '../organization.module.css';

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

  const values = watch(['emails', 'type']);

  const onSubmit = useCallback(
    async ({ emails, type, team }: InviteSchemaType) => {
      const emailList = emails
        .split(',')
        .map(e => e.trim())
        .filter(str => str.length > 0);

      if (!organization?.id) return;
      if (!emailList.length) return;
      if (!type) return;

      try {
        await client?.frontierServiceCreateOrganizationInvitation(
          organization?.id,
          {
            userIds: emailList,
            groupIds: team ? [team] : undefined,
            roleIds: [type]
          }
        );
        toast.success('members added');

        navigate({ to: '/members' });
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    },
    [client, navigate, organization?.id]
  );

  useEffect(() => {
    async function getInformation() {
      try {
        setIsLoading(true);

        if (!organization?.id) return;
        
        const orgRolesRes = await client?.frontierServiceListOrganizationRoles(
          organization.id,
          {
            scopes: [PERMISSIONS.OrganizationNamespace]
          }
        );
        const orgRoles = orgRolesRes?.data?.roles ?? []
        
        const serviceListRolesRes = await client?.frontierServiceListRoles({
            scopes: [PERMISSIONS.OrganizationNamespace]
        });
        const roles = serviceListRolesRes?.data?.roles ?? []


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
    const [emails, type] = values;
    const emailList =
      emails
        ?.split(',')
        .map((e: string) => e.trim())
        .filter(str => str.length > 0) || [];
    return emailList.length <= 0 || !type || isSubmitting;
  }, [isSubmitting, values]);

  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%', zIndex: '60' }}
        overlayClassname={styles.overlay}
      >
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex justify="between" style={{ padding: '16px 24px' }}>
            <Text size="large weight="medium">
              Invite people
            </Text>

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              src={cross as unknown as string}
              onClick={() => navigate({ to: '/members' })}
              data-test-id="frontier-sdk-invite-member-close-btn"
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
                    placeholder="Enter comma separated emails like abc@domain.com, bcd@domain.com"
                  />
                )}
                control={control}
                name="emails"
              />

              <Text size="mini" variant="danger">
                {errors.emails && String(errors.emails?.message)}
              </Text>
            </InputField>
            <InputField label="Invite as">
              {isLoading ? (
                <Skeleton height={'25px'} />
              ) : (
                <Controller
                  render={({ field }) => {
                    const { ref, onChange, ...rest } = field;
                    return (
                      <Select {...rest} onValueChange={onChange}>
                        <Select.Trigger ref={ref}>
                          <Select.Value placeholder="Select a role" />
                        </Select.Trigger>
                        <Select.Content
                          style={{
                            zIndex: 65
                          }}
                        >
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
                    );
                  }}
                  control={control}
                  name="type"
                />
              )}
              <Text size="mini" variant="danger">
                {errors.type && String(errors.type?.message)}
              </Text>
            </InputField>

            <InputField label="Add to team (optional)">
              {isLoading ? (
                <Skeleton height={'25px'} />
              ) : (
                <Controller
                  render={({ field }) => {
                    const { ref, onChange, ...rest } = field;
                    return (
                      <Select {...rest} onValueChange={onChange}>
                        <Select.Trigger ref={ref}>
                          <Select.Value placeholder="Select a team" />
                        </Select.Trigger>
                        <Select.Content
                          style={{
                            zIndex: 65
                          }}
                        >
                          <Select.Group>
                            {!teams.length && (
                              <Text className={styles.noSelectItem}>
                                No teams available
                              </Text>
                            )}
                            {teams.map(t => (
                              <Select.Item value={t.id} key={t.id}>
                                {t.title}
                              </Select.Item>
                            ))}
                          </Select.Group>
                        </Select.Content>
                      </Select>
                    );
                  }}
                  control={control}
                  name="team"
                />
              )}
              <Text size="mini" variant="danger">
                {errors.team && String(errors.team?.message)}
              </Text>
            </InputField>
            <Separator />
            <Flex justify="end">
              <Button
                type="submit"
                disabled={isDisabled}
                data-test-id="frontier-sdk-send-member-invite-btn"
                loading={isSubmitting}
                loaderText="Sending..."
              >
                Send invite
              </Button>
            </Flex>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
