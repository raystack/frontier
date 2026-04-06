'use client';

import { useEffect, useMemo, ReactNode } from 'react';
import {
  AlertDialog,
  Button,
  Dialog,
  Flex,
  IconButton,
  Image,
  Skeleton,
  Text,
  toastManager
} from '@raystack/apsara-v1';
import { CheckCircledIcon } from '@radix-ui/react-icons';
import { type Domain } from '@raystack/proton/frontier';
import { useFrontier } from '../../contexts/FrontierContext';
import { usePermissions } from '../../hooks/usePermissions';
import { useOrganizationDomains } from '../../hooks/useOrganizationDomains';
import { PERMISSIONS, shouldShowComponent } from '../../../utils';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import { AddDomainDialog } from './components/add-domain-dialog';
import { VerifyDomainDialog, type VerifyDomainPayload } from './components/verify-domain-dialog';
import { DeleteDomainDialog, type DeleteDomainPayload } from './components/delete-domain-dialog';
import styles from './security-view.module.css';
import deleteIcon from '../../assets/delete.svg';

const addDomainDialogHandle = Dialog.createHandle();
const verifyDomainDialogHandle = Dialog.createHandle<VerifyDomainPayload>();
const deleteDomainDialogHandle = AlertDialog.createHandle<DeleteDomainPayload>();

export interface SecurityViewProps {
  children?: ReactNode;
}

export function SecurityView({ children }: SecurityViewProps) {
  const { activeOrganization: organization } = useFrontier();
  const { isFetching, domains, refetch, error: domainsError } = useOrganizationDomains();

  useEffect(() => {
    if (domainsError) {
      toastManager.add({
        title: 'Something went wrong',
        description: (domainsError as Error).message,
        type: 'error'
      });
    }
  }, [domainsError]);

  const resource = `app/organization:${organization?.id}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.UpdatePermission,
        resource
      }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!organization?.id
  );

  const canUpdateOrg = shouldShowComponent(
    permissions,
    `${PERMISSIONS.UpdatePermission}::${resource}`
  );

  const isLoading = !organization?.id || isFetching || isPermissionsFetching;

  return (
    <ViewContainer>
      <ViewHeader
        title="Security"
        description="Manage security for this domain."
      />

      <Flex direction="column">
        <Flex direction="column" gap={7} className={styles.section}>
          <Flex align="center" gap={9}>
            <Flex direction="column" gap={3} className={styles.sectionHeader}>
              {isLoading ? (
                <>
                  <Skeleton width="30%" height={24} />
                  <Skeleton width="60%" height={16} />
                </>
              ) : (
                <>
                  <Text size="large" weight="medium">
                    Allowed email domains
                  </Text>
                  <Text size="small" variant="secondary">
                    Anyone with an email address at these domains is allowed to sign up for this workspace.
                  </Text>
                </>
              )}
            </Flex>
            {isLoading ? (
              <Skeleton width={110} height={33} />
            ) : canUpdateOrg ? (
              <Button
                variant="outline"
                color="neutral"
                size="normal"
                onClick={() => addDomainDialogHandle.open(null)}
                data-test-id="frontier-sdk-add-domain-btn"
              >
                Add domain
              </Button>
            ) : null}
          </Flex>
          {isLoading ? (
            <Skeleton height={52} />
          ) : domains.length > 0 ? domains.map((domain) => (
            <DomainCard
              key={domain.id}
              domain={domain}
              canUpdateOrg={canUpdateOrg}
              onVerify={(domainId) =>
                verifyDomainDialogHandle.openWithPayload({ domainId })
              }
              onDelete={(domainId) =>
                deleteDomainDialogHandle.openWithPayload({ domainId })
              }
            />
          )) : null}
        </Flex>
        {children}
      </Flex>

      <AddDomainDialog
        handle={addDomainDialogHandle}
        onDomainAdded={(domainId) => {
          addDomainDialogHandle.close();
          verifyDomainDialogHandle.openWithPayload({ domainId });
          refetch();
        }}
      />

      <VerifyDomainDialog
        handle={verifyDomainDialogHandle}
        refetch={refetch}
      />

      <DeleteDomainDialog
        handle={deleteDomainDialogHandle}
        refetch={refetch}
      />
    </ViewContainer >
  );
}

interface DomainsListProps {
  domain: Domain;
  canUpdateOrg: boolean;
  onVerify: (domainId: string) => void;
  onDelete: (domainId: string) => void;
}

function DomainCard({ domain, canUpdateOrg, onVerify, onDelete }: DomainsListProps) {
  return (
    <Flex className={styles.domainCell} justify="between" align="center">
      <Text size="regular">{domain.name}</Text>
      <Flex gap={7} align="center" justify="end">
        {domain.state === 'verified' ? (<Flex gap={1} align="center" className={styles.verified}>
          <CheckCircledIcon />
          <Text size="mini" weight="medium" className={styles.verified}>Verified</Text>
        </Flex>
        ) : canUpdateOrg ? (
          <Button
            variant="outline"
            color="neutral"
            size="small"
            onClick={() => onVerify(domain.id || '')}
            data-test-id="frontier-sdk-verify-domain-btn"
          >
            Verify
          </Button>
        ) : null}
        {canUpdateOrg ? (
          <IconButton
            size={3}
            onClick={() => onDelete(domain.id || '')}
            data-test-id="frontier-sdk-delete-domain-btn"
          >
            <Image src={deleteIcon as unknown as string} alt="Delete" width={16} height={16} />
          </IconButton>
        ) : null}
      </Flex>
    </Flex>
  )
}
