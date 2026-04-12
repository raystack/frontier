'use client';

import {
    Button,
    Separator,
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
import { useMutation } from '@connectrpc/connect-query';
import {
    FrontierServiceQueries,
    UpdateProjectRequestSchema,
    Project,
    Organization
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { AuthTooltipMessage } from '~/react/utils';
import { handleConnectError } from '~/utils/error';

const projectSchema = yup
    .object({
        title: yup.string().required(),
        name: yup.string().required()
    })
    .required();

type FormData = yup.InferType<typeof projectSchema>;

interface ProjectGeneralProps {
    projectId: string;
    project?: Project;
    organization?: Organization;
    isLoading?: boolean;
    onDeleteClick?: () => void;
}

export const ProjectGeneral = ({
    projectId,
    organization,
    project,
    isLoading: isProjectLoading,
    onDeleteClick
}: ProjectGeneralProps) => {
    const {
        reset,
        handleSubmit,
        formState: { errors, isSubmitting },
        register
    } = useForm({
        resolver: yupResolver(projectSchema)
    });

    const { mutateAsync: updateProject } = useMutation(
        FrontierServiceQueries.updateProject,
        {
            onSuccess: () => toast.success('Project updated successfully')
        }
    );

    useEffect(() => {
        reset(project);
    }, [reset, project]);

    const resource = `app/project:${projectId}`;
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
        !!projectId
    );

    const { canUpdateProject, canDeleteProject } = useMemo(() => {
        return {
            canUpdateProject: shouldShowComponent(
                permissions,
                `${PERMISSIONS.UpdatePermission}::${resource}`
            ),
            canDeleteProject: shouldShowComponent(
                permissions,
                `${PERMISSIONS.DeletePermission}::${resource}`
            )
        };
    }, [permissions, resource]);

    async function onSubmit(data: FormData) {
        if (!organization?.id) return;
        if (!projectId) return;

        try {
            await updateProject(
                create(UpdateProjectRequestSchema, {
                    id: projectId,
                    body: {
                        name: data.name,
                        title: data.title,
                        orgId: organization.id
                    }
                })
            );
        } catch (error) {
            handleConnectError(error, {
                AlreadyExists: () => toast.error('Project already exists'),
                PermissionDenied: () => toast.error('You don\'t have permission to perform this action'),
                InvalidArgument: (err) => toast.error('Invalid input', { description: err.message }),
                NotFound: (err) => toast.error('Not found', { description: err.message }),
                Default: (err) => toast.error('Something went wrong', { description: err.message }),
            });
        }
    }

    const isLoading = isPermissionsFetching || isProjectLoading;

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
                            label="Project title"
                            size="large"
                            error={errors.title && String(errors.title?.message)}
                            {...register('title')}
                            placeholder="Provide project title"
                        />
                    )}
                    {isLoading ? (
                        <div>
                            <Skeleton height={'16px'} />
                            <Skeleton height={'32px'} />
                        </div>
                    ) : (
                        <InputField
                            label="Project name"
                            size="large"
                            error={errors.name && String(errors.name?.message)}
                            {...register('name')}
                            disabled
                            placeholder="Provide project name"
                        />
                    )}
                    {isLoading ? (
                        <Skeleton height={'32px'} width={'64px'} />
                    ) : (
                        <Tooltip message={AuthTooltipMessage} disabled={canUpdateProject}>
                            <Button
                                type="submit"
                                disabled={!canUpdateProject}
                                data-test-id="frontier-sdk-update-project-btn"
                                loading={isSubmitting}
                                loaderText="Updating..."
                            >
                                Update project
                            </Button>
                        </Tooltip>
                    )}
                </Flex>
            </form>
            <Separator />

            <GeneralDeleteProject
                canDeleteProject={canDeleteProject}
                isLoading={isLoading}
                onDeleteClick={onDeleteClick}
            />
        </Flex>
    );
};

interface GeneralDeleteProjectProps {
    canDeleteProject?: boolean;
    isLoading?: boolean;
    onDeleteClick?: () => void;
}

const GeneralDeleteProject = ({
    canDeleteProject,
    isLoading,
    onDeleteClick
}: GeneralDeleteProjectProps) => {
    return (
        <Flex direction="column" gap={5}>
            {isLoading ? (
                <Skeleton height={'16px'} width={'50%'} />
            ) : (
                <Text size={3} variant="secondary">
                    If you want to permanently delete this project and all of its data.
                </Text>
            )}{' '}
            {isLoading ? (
                <Skeleton height={'32px'} width={'64px'} />
            ) : (
                <Tooltip message={AuthTooltipMessage} disabled={canDeleteProject}>
                    <Button
                        variant="solid"
                        color="danger"
                        type="submit"
                        data-test-id="frontier-sdk-delete-project-btn"
                        disabled={!canDeleteProject}
                        onClick={onDeleteClick}
                    >
                        Delete project
                    </Button>
                </Tooltip>
            )}
        </Flex>
    );
};

