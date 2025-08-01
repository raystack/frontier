import {
  Button,
  Separator,
  toast,
  Tooltip,
  Skeleton,
  Text,
  Flex,
  InputField
} from '@raystack/apsara/v1';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useEffect, useMemo } from 'react';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { usePermissions } from '~/react/hooks/usePermissions';
import {
  V1Beta1Organization,
  V1Beta1Project,
  V1Beta1ProjectRequestBody
} from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { AuthTooltipMessage } from '~/react/utils';

const projectSchema = yup
  .object({
    title: yup.string().required(),
    name: yup.string().required()
  })
  .required();

type FormData = yup.InferType<typeof projectSchema>;

interface GeneralProjectProps {
  project?: V1Beta1Project;
  organization?: V1Beta1Organization;
  isLoading?: boolean;
}

export const General = ({
  organization,
  project,
  isLoading: isProjectLoading
}: GeneralProjectProps) => {
  const {
    reset,
    control,
    handleSubmit,
    formState: { errors, isSubmitting },
    register
  } = useForm({
    resolver: yupResolver(projectSchema)
  });
  let { projectId } = useParams({ from: '/projects/$projectId' });
  const { client } = useFrontier();

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
    if (!client) return;
    if (!organization?.id) return;
    if (!projectId) return;

    try {
      const payload: V1Beta1ProjectRequestBody = {
        name: data.name,
        title: data.title,
        orgId: organization.id
      };
      await client.frontierServiceUpdateProject(projectId, payload);
      toast.success('Project updated');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
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
        organization={organization}
        canDeleteProject={canDeleteProject}
        isLoading={isLoading}
      />
    </Flex>
  );
};

interface GeneralDeleteProjectProps extends GeneralProjectProps {
  canDeleteProject?: boolean;
}

export const GeneralDeleteProject = ({
  canDeleteProject,
  isLoading
}: GeneralDeleteProjectProps) => {
  let { projectId } = useParams({ from: '/projects/$projectId' });
  const navigate = useNavigate({ from: '/projects/$projectId' });

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
            onClick={() =>
              navigate({
                to: `/projects/$projectId/delete`,
                params: { projectId: projectId }
              })
            }
          >
            Delete project
          </Button>
        </Tooltip>
      )}
    </Flex>
  );
};
