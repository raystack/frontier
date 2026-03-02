'use client';

import { useMemo, useState } from 'react';
import {
  Button,
  Tooltip,
  Separator,
  Skeleton,
  Text,
  Flex
} from '@raystack/apsara';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { usePermissions } from '~/react/hooks/usePermissions';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { GeneralOrganization } from './general-organization';
import { AuthTooltipMessage } from '~/react/utils';
import { useTerminology } from '~/react/hooks/useTerminology';
import { PageHeader } from '~/react/components/common/page-header';
import { DeleteOrganizationDialog } from './delete-organization-dialog';
import sharedStyles from '../../components/organization/styles.module.css';

export interface GeneralPageProps {
  onDeleteSuccess?: () => void;
}

export function GeneralPage({ onDeleteSuccess }: GeneralPageProps = {}) {
  const t = useTerminology();
  const { activeOrganization: organization, isActiveOrganizationLoading } =
    useFrontier();

  const resource = `app/organization:${organization?.id}`;

  const listOfPermissionsToCheck = useMemo(() => {
    return [
      {
        permission: PERMISSIONS.UpdatePermission,
        resource: resource
      },
      {
        permission: PERMISSIONS.DeletePermission,
        resource: resource
      }
    ];
  }, [resource]);

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!organization?.id
  );

  const { canUpdateWorkspace, canDeleteWorkspace } = useMemo(() => {
    return {
      canUpdateWorkspace: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      ),
      canDeleteWorkspace: shouldShowComponent(
        permissions,
        `${PERMISSIONS.DeletePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const isLoading = isActiveOrganizationLoading || isPermissionsFetching;

  const [showDeleteDialog, setShowDeleteDialog] = useState(false);

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex direction="column" className={sharedStyles.container}>
        <Flex
          direction="row"
          justify="between"
          align="center"
          className={sharedStyles.header}
        >
          <PageHeader
            title="General"
            description={`Basic configuration for the ${t.organization({
              case: 'lower'
            })}.`}
          />
        </Flex>
        <Flex direction="column" gap={9}>
          <GeneralOrganization
            organization={organization}
            canUpdateWorkspace={canUpdateWorkspace}
            isLoading={isLoading}
          />
          <Separator />
          <Flex direction="column" gap={5}>
            {isLoading ? (
              <Skeleton height={'16px'} width={'50%'} />
            ) : (
              <Text size={3} variant="secondary">
                If you want to permanently delete this{' '}
                {t.organization({ case: 'lower' })} and all of its data.
              </Text>
            )}
            {isLoading ? (
              <Skeleton height={'32px'} width={'64px'} />
            ) : (
              <Tooltip disabled={canDeleteWorkspace} message={AuthTooltipMessage}>
                <Button
                  variant="solid"
                  color="danger"
                  type="submit"
                  onClick={() => setShowDeleteDialog(true)}
                  disabled={!canDeleteWorkspace}
                  data-test-id="frontier-sdk-delete-organization-btn"
                >
                  Delete {t.organization({ case: 'lower' })}
                </Button>
              </Tooltip>
            )}
            <DeleteOrganizationDialog
              open={showDeleteDialog}
              onOpenChange={setShowDeleteDialog}
              onDeleteSuccess={onDeleteSuccess}
            />
          </Flex>
        </Flex>
      </Flex>
    </Flex>
  );
}

