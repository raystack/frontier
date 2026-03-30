'use client';

import { forwardRef, useCallback, useImperativeHandle, useMemo, useState } from 'react';
import { yupResolver } from '@hookform/resolvers/yup';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import { create } from '@bufbuild/protobuf';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  CreateOrganizationInvitationRequestSchema,
  ListOrganizationRolesRequestSchema,
  ListRolesRequestSchema,
  ListOrganizationGroupsRequestSchema
} from '@raystack/proton/frontier';
import {
  Button,
  Skeleton,
  Text,
  Label,
  Select,
  Flex,
  Dialog,
  TextArea,
  toastManager
} from '@raystack/apsara-v1';
import { useFrontier } from '../../../contexts/FrontierContext';
import { PERMISSIONS } from '../../../../utils';

const inviteSchema = yup.object({
  type: yup.string().required(),
  team: yup.string(),
  emails: yup.string().required()
});

type InviteSchemaType = yup.InferType<typeof inviteSchema>;

export interface InviteMemberDialogHandle {
  open: () => void;
}

export interface InviteMemberDialogProps {
  showTeamField?: boolean;
  refetch: () => void;
}

export const InviteMemberDialog = forwardRef<
  InviteMemberDialogHandle,
  InviteMemberDialogProps
>(function InviteMemberDialog({ showTeamField = true, refetch }, ref) {
  const [isOpen, setIsOpen] = useState(false);
  const {
    watch,
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(inviteSchema)
  });
  const { activeOrganization: organization } = useFrontier();

  useImperativeHandle(ref, () => ({
    open: () => setIsOpen(true)
  }));

  const handleOpenChange = (value: boolean) => {
    setIsOpen(value);
    if (!value) {
      reset();
      refetch();
    }
  };

  const { data: orgRolesData, isLoading: isOrgRolesLoading } = useQuery(
    FrontierServiceQueries.listOrganizationRoles,
    create(ListOrganizationRolesRequestSchema, {
      orgId: organization?.id || '',
      scopes: [PERMISSIONS.OrganizationNamespace]
    }),
    { enabled: !!organization?.id }
  );

  const orgRoles = useMemo(() => orgRolesData?.roles || [], [orgRolesData]);

  const { data: globalRolesData, isLoading: isGlobalRolesLoading } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, {
      scopes: [PERMISSIONS.OrganizationNamespace]
    }),
    { enabled: !!organization?.id }
  );

  const globalRoles = useMemo(
    () => globalRolesData?.roles || [],
    [globalRolesData]
  );

  const { data: teamsData, isLoading: isGroupsLoading } = useQuery(
    FrontierServiceQueries.listOrganizationGroups,
    create(ListOrganizationGroupsRequestSchema, {
      orgId: organization?.id || ''
    }),
    { enabled: !!organization?.id && showTeamField }
  );

  const teams = useMemo(() => teamsData?.groups || [], [teamsData]);

  const isLoading =
    isOrgRolesLoading || isGlobalRolesLoading || (showTeamField && isGroupsLoading);

  const roles = useMemo(
    () => [...(globalRoles || []), ...(orgRoles || [])],
    [globalRoles, orgRoles]
  );

  const { mutateAsync: createInvitation } = useMutation(
    FrontierServiceQueries.createOrganizationInvitation,
    {
      onSuccess: () => {
        toastManager.add({ title: 'User(s) invited', type: 'success' });
        handleOpenChange(false);
      },
      onError: (error: Error) => {
        toastManager.add({
          title: 'Something went wrong',
          description: error?.message || 'Failed to create invitation',
          type: 'error'
        });
      }
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
          groupIds: showTeamField && team ? [team] : undefined,
          roleIds: [type]
        });
        await createInvitation(req);
      } catch (error: unknown) {
        toastManager.add({
          title: 'Something went wrong',
          description:
            error instanceof Error
              ? error.message
              : 'Failed to create invitation',
          type: 'error'
        });
      }
    },
    [createInvitation, organization?.id, showTeamField]
  );

  const isDisabled = useMemo(() => {
    const [emails, type] = values;
    const emailList =
      emails
        ?.split(',')
        .map((e: string) => e.trim())
        .filter((str: string) => str.length > 0) || [];
    return emailList.length <= 0 || !type || isSubmitting;
  }, [isSubmitting, values]);

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <Dialog.Content width={600}>
        <Dialog.Header>
          <Dialog.Title>Invite people</Dialog.Title>
        </Dialog.Header>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Body>
            <Flex direction="column" gap={7}>
              {isLoading ? (
                <Skeleton height="80px" />
              ) : (
                <TextArea
                  label="Email"
                  {...register('emails')}
                  placeholder="abc@example.com, xyz@example.com"
                />
              )}
              <Flex direction="column" gap={2}>
                <Label>Invite as</Label>
                {isLoading ? (
                  <Skeleton height="36px" />
                ) : (
                  <Controller
                    render={({ field }) => {
                      const { ref, onChange, ...rest } = field;
                      return (
                        <Select
                          {...rest}
                          onValueChange={onChange}
                        >
                          <Select.Trigger ref={ref}>
                            <Select.Value placeholder="Select a role" />
                          </Select.Trigger>
                          <Select.Content>
                            {!roles.length && (
                              <Text size="small" variant="secondary">
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
                          </Select.Content>
                        </Select>
                      );
                    }}
                    control={control}
                    name="type"
                  />
                )}
              </Flex>
              {showTeamField && (
                <Flex direction="column" gap={2}>
                  <Label>Add to team (optional)</Label>
                  {isLoading ? (
                    <Skeleton height="36px" />
                  ) : (
                    <Controller
                      render={({ field }) => {
                        const { ref, onChange, ...rest } = field;
                        return (
                          <Select
                            {...rest}
                            onValueChange={onChange}
                          >
                            <Select.Trigger ref={ref}>
                              <Select.Value placeholder="Select a team" />
                            </Select.Trigger>
                            <Select.Content>
                              {!teams?.length && (
                                <Text size="small" variant="secondary">
                                  No teams available
                                </Text>
                              )}
                              {(teams || []).map(t => (
                                <Select.Item value={t.id || ''} key={t.id}>
                                  {t.title}
                                </Select.Item>
                              ))}
                            </Select.Content>
                          </Select>
                        );
                      }}
                      control={control}
                      name="team"
                    />
                  )}
                </Flex>
              )}
            </Flex>
          </Dialog.Body>
          <Dialog.Footer>
            <Flex justify="end">
              <Button
                type="submit"
                variant="solid"
                color="accent"
                disabled={isDisabled}
                data-test-id="frontier-sdk-send-member-invite-btn"
                loading={isSubmitting}
                loaderText="Sending..."
              >
                Send invites
              </Button>
            </Flex>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
});
