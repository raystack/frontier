'use client';

import {
    Separator,
    Button,
    toast,
    Tooltip,
    Skeleton,
    Text,
    Flex,
    InputField
} from '@raystack/apsara';

import { yupResolver } from '@hookform/resolvers/yup';
import { useEffect, useMemo } from 'react';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { usePermissions } from '~/react/hooks/usePermissions';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { AuthTooltipMessage } from '~/react/utils';
import { useMutation } from '@connectrpc/connect-query';
import {
    FrontierServiceQueries,
    UpdateGroupRequestSchema,
    type Group,
    type Organization
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { handleConnectError } from '~/utils/error';

const teamSchema = yup
    .object({
        title: yup.string().required(),
        name: yup.string().required()
    })
    .required();

type FormData = yup.InferType<typeof teamSchema>;

interface TeamGeneralProps {
    team?: Group;
    organization?: Organization;
    teamId: string;
    isLoading?: boolean;
    onDeleteClick?: () => void;
}

export const TeamGeneral = ({
    organization,
    team,
    teamId,
    isLoading: isTeamLoading,
    onDeleteClick
}: TeamGeneralProps) => {
    const {
        reset,
        handleSubmit,
        formState: { errors, isSubmitting },
        register
    } = useForm({
        resolver: yupResolver(teamSchema)
    });

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

    const { mutateAsync: updateTeam } = useMutation(
        FrontierServiceQueries.updateGroup,
        {
            onSuccess: () => {
                toast.success('Team updated');
            }
        }
    );

    async function onSubmit(data: FormData) {
        if (!organization?.id) return;
        if (!teamId) return;

        const request = create(UpdateGroupRequestSchema, {
            id: teamId,
            orgId: organization.id,
            body: {
                title: data.title,
                name: data.name
            }
        });

        try {
            await updateTeam(request);
        } catch (error) {
            handleConnectError(error, {
                AlreadyExists: () => toast.error('Team already exists'),
                PermissionDenied: () => toast.error('You don\'t have permission to perform this action'),
                InvalidArgument: (err) => toast.error('Invalid input', { description: err.message }),
                NotFound: (err) => toast.error('Not found', { description: err.message }),
                Default: (err) => toast.error('Something went wrong', { description: err.message }),
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
                            onClick={onDeleteClick}
                            data-test-id="frontier-sdk-delete-team-btn"
                        >
                            Delete team
                        </Button>
                    </Tooltip>
                )}
            </Flex>
        </Flex>
    );
};

