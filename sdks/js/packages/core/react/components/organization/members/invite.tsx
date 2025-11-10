import {
  Button,
  toast,
  Skeleton,
  Image,
  Text,
  Select,
  Flex,
  Dialog,
  TextArea,
  Label
} from '@raystack/apsara';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate } from '@tanstack/react-router';
import { useCallback, useMemo } from 'react';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { PERMISSIONS } from '~/utils';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries, CreateOrganizationInvitationRequestSchema, ListOrganizationRolesRequestSchema, ListRolesRequestSchema, ListOrganizationGroupsRequestSchema } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { handleSelectValueChange } from '~/react/utils';
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
    register,
    control,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(inviteSchema)
  });
  const navigate = useNavigate({ from: '/members/modal' });
  const { activeOrganization: organization } = useFrontier();
  
  // Organization roles query
  const { data: orgRoles, isLoading: isOrgRolesLoading, error: orgRolesError } = useQuery(
    FrontierServiceQueries.listOrganizationRoles,
    create(ListOrganizationRolesRequestSchema, {
      orgId: organization?.id || '',
      scopes: [PERMISSIONS.OrganizationNamespace]
    }),
    { 
      enabled: !!organization?.id,
      select: (data) => data?.roles || []
    }
  );
  
  // Global roles query
  const { data: globalRoles, isLoading: isGlobalRolesLoading, error: globalRolesError } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, {
      scopes: [PERMISSIONS.OrganizationNamespace]
    }),
    { 
      enabled: !!organization?.id,
      select: (data) => data?.roles || []
    }
  );
  
  // Organization groups query
  const { data: teams, isLoading: isGroupsLoading, error: groupsError } = useQuery(
    FrontierServiceQueries.listOrganizationGroups,
    create(ListOrganizationGroupsRequestSchema, {
      orgId: organization?.id || ''
    }),
    { 
      enabled: !!organization?.id,
      select: (data) => data?.groups || []
    }
  );
  
  const isLoading = isOrgRolesLoading || isGlobalRolesLoading || isGroupsLoading;
  
  const roles = useMemo(() => 
    [...(globalRoles || []), ...(orgRoles || [])],
    [globalRoles, orgRoles]
  );
  
  const { mutateAsync: createInvitation } = useMutation(
    FrontierServiceQueries.createOrganizationInvitation,
    {
      onSuccess: () => {
        toast.success('User(s) invited');
        navigate({ to: '/members' });
      },
      onError: (error: any) => {
        toast.error('Something went wrong', {
          description: error?.message || 'Failed to create invitation'
        });
      },
    }
  );

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
        const req = create(CreateOrganizationInvitationRequestSchema, {
          orgId: organization.id,
          userIds: emailList,
          groupIds: team ? [team] : undefined,
          roleIds: [type]
        });
        await createInvitation(req);
      } catch (error: any) {
        toast.error('Something went wrong', {
          description: error?.message || 'Failed to create invitation'
        });
      }
    },
    [createInvitation, organization?.id]
  );



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
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%' }}
        overlayClassName={styles.overlay}
      >
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="large" weight="medium">
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
        </Dialog.Header>
        <Dialog.Body>
          <form onSubmit={handleSubmit(onSubmit)}>
            <Flex direction="column" gap={5}>
              {isLoading ? (
                <Skeleton height="52px" />
              ) : (
                <TextArea
                  label="Email"
                  required
                  {...register('emails')}
                  //  TODO: Fix placeholder prop in apsara TextAreaProps
                  //  @ts-expect-error placeholder prop exists on textarea but not in apsara TextAreaProps type definition
                  placeholder="Enter comma separated emails like abc@domain.com, bcd@domain.com"
                  error={Boolean(errors.emails?.message)}
                  helperText={
                    errors.emails?.message ? String(errors.emails?.message) : ''
                  }
                />
              )}
              <Flex direction="column" gap={2}>
                <Label>Invite as</Label>
                {isLoading ? (
                  <Skeleton height={'25px'} />
                ) : (
                  <Controller
                    render={({ field }) => {
                      const { ref, onChange, ...rest } = field;
                      return (
                        <Select
                          {...rest}
                          onValueChange={handleSelectValueChange(onChange)}
                        >
                          <Select.Trigger ref={ref}>
                            <Select.Value placeholder="Select a role" />
                          </Select.Trigger>
                          <Select.Content>
                            <Select.Group>
                              {!roles.length && (
                                <Text className={styles.noSelectItem}>
                                  No roles available
                                </Text>
                              )}
                              {roles.map(role => (
                                <Select.Item
                                  value={role.id || ''}
                                  key={role.id}
                                >
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
              </Flex>
              <Flex direction="column" gap={2}>
                <Label>Add to team (optional)</Label>
                {isLoading ? (
                  <Skeleton height={'25px'} />
                ) : (
                  <Controller
                    render={({ field }) => {
                      const { ref, onChange, ...rest } = field;
                      return (
                        <Select
                          {...rest}
                          onValueChange={handleSelectValueChange(onChange)}
                        >
                          <Select.Trigger ref={ref}>
                            <Select.Value placeholder="Select a team" />
                          </Select.Trigger>
                          <Select.Content>
                            <Select.Group>
                              {!teams?.length && (
                                <Text className={styles.noSelectItem}>
                                  No teams available
                                </Text>
                              )}
                              {(teams || []).map(t => (
                                <Select.Item value={t.id || ''} key={t.id}>
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
              </Flex>
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
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
