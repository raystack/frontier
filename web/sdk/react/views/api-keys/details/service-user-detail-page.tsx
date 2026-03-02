import { useState } from 'react';
import { Button, Flex, Text, Skeleton, Image } from '@raystack/apsara';
import { PageHeader } from '~/react/components/common/page-header';
import { useFrontier } from '~/react/contexts/FrontierContext';
import AddServiceUserToken from './add-service-user-token';
import { CheckCircledIcon, CopyIcon } from '@radix-ui/react-icons';
import { useCopyToClipboard } from '~/react/hooks/useCopyToClipboard';
import { useTerminology } from '~/react/hooks/useTerminology';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import {
  FrontierServiceQueries,
  GetServiceUserRequestSchema,
  type ServiceUserToken
} from '@raystack/proton/frontier';
import { useServiceUserTokens } from '../hooks/useServiceUserTokens';
import ManageServiceUserProjectsDialog from './manage-service-user-projects-dialog';
import { DeleteServiceUserKeyDialog } from './delete-service-user-key-dialog';

import backIcon from '~/react/assets/chevron-left.svg';
import sharedStyles from '../../../components/organization/styles.module.css';
import styles from './service-user.module.css';

const Headings = ({
  isLoading,
  name,
  onManageAccessClick
}: {
  isLoading: boolean;
  name: string;
  onManageAccessClick?: () => void;
}) => {
  const t = useTerminology();

  return (
    <Flex direction="column" gap={3} style={{ width: '100%' }}>
      {isLoading ? (
        <Skeleton containerClassName={styles.flex1} />
      ) : (
        <Flex justify="between">
          <Text size="large">{name}</Text>
          <Button
            variant="outline"
            color="neutral"
            onClick={onManageAccessClick}
            data-test-id="frontier-sdk-service-account-manage-access-open-btn"
          >
            Manage access
          </Button>
        </Flex>
      )}
      {isLoading ? (
        <Skeleton containerClassName={styles.flex1} />
      ) : (
        <Text size="regular" variant="secondary">
          Create API key for accessing {t.appName()} and its features
        </Text>
      )}
    </Flex>
  );
};

const ServiceUserTokenItem = ({
  token,
  isLoading,
  onRevokeClick
}: {
  token: ServiceUserToken;
  isLoading: boolean;
  onRevokeClick?: (tokenId: string) => void;
}) => {
  const [isCopied, setIsCopied] = useState(false);
  const { copy } = useCopyToClipboard();

  const encodedToken = 'Basic ' + btoa(`${token?.id}:${token?.token}`);

  async function onCopy() {
    const res = await copy(encodedToken);
    if (res) {
      setIsCopied(true);
      setTimeout(() => {
        setIsCopied(false);
      }, 1000);
    }
  }

  return (
    <Flex className={styles.serviceKeyItem} direction="column" gap={3}>
      <Flex justify="between" style={{ width: '100%' }} align="center">
        {isLoading ? (
          <>
            <Skeleton containerClassName={styles.flex1} width={'300px'} />
            <Skeleton containerClassName={styles.serviceKeyItemLoaderBtn} />
          </>
        ) : (
          <>
            <Text size={3} weight="medium">
              {token?.title}
            </Text>
            <Button
              variant="outline"
              color="neutral"
              size={'small'}
              data-test-id={`frontier-sdk-service-account-token-revoke-btn`}
              onClick={() => onRevokeClick?.(token?.id || '')}
            >
              Revoke
            </Button>
          </>
        )}
      </Flex>
      {token?.token ? (
        <Flex gap={3} direction="column">
          <Text size="small" variant="secondary" weight="regular">
            Note: Please save your key securely, it cannot be recovered after
            leaving this page
          </Text>
          <Flex className={styles.tokenBox} justify="between" gap={5}>
            <Text size="small" weight="medium" className={styles.tokenText}>
              {encodedToken}
            </Text>
            {isCopied ? (
              <CheckCircledIcon color="var(--rs-color-foreground-success-primary)" />
            ) : (
              <CopyIcon
                onClick={onCopy}
                data-test-id={`frontier-sdk-service-account-token-copy-btn`}
                style={{ cursor: 'pointer' }}
              />
            )}
          </Flex>
        </Flex>
      ) : null}
    </Flex>
  );
};

const SerivceUserTokenList = ({
  isLoading,
  tokens,
  onRevokeClick
}: {
  isLoading: boolean;
  tokens: ServiceUserToken[];
  onRevokeClick?: (tokenId: string) => void;
}) => {
  const tokenList = isLoading
    ? [
        ...new Array(3).map(
          (_, i) => ({ id: i.toString() } as ServiceUserToken)
        )
      ]
    : tokens;

  return (
    <Flex direction="column" className={styles.serviceKeyList}>
      {tokenList.map(token => (
        <ServiceUserTokenItem
          token={token}
          key={token?.id}
          isLoading={isLoading}
          onRevokeClick={onRevokeClick}
        />
      ))}
    </Flex>
  );
};

export interface ServiceUserDetailPageProps {
  serviceUserId: string;
  onBack?: () => void;
  enableTokensFetch?: boolean;
}

export default function ServiceUserDetailPage({
  serviceUserId,
  onBack,
  enableTokensFetch = true
}: ServiceUserDetailPageProps) {
  const { activeOrganization } = useFrontier();
  const orgId = activeOrganization?.id || '';

  const { data: serviceUser, isLoading: isServiceUserLoading } = useQuery(
    FrontierServiceQueries.getServiceUser,
    create(GetServiceUserRequestSchema, {
      id: serviceUserId,
      orgId
    }),
    {
      enabled: Boolean(serviceUserId) && Boolean(orgId),
      select: data => data?.serviceuser
    }
  );

  const {
    tokens: serviceUserTokens,
    isLoading: isServiceUserTokensLoading,
    addToken: onAddToken
  } = useServiceUserTokens({
    id: serviceUserId,
    orgId,
    enableFetch: enableTokensFetch
  });

  const isLoading = isServiceUserLoading || isServiceUserTokensLoading;

  const [showProjectsDialog, setShowProjectsDialog] = useState(false);

  const [deleteKeyState, setDeleteKeyState] = useState({
    open: false,
    tokenId: ''
  });

  const handleDeleteKeyOpenChange = (value: boolean) => {
    if (!value) {
      setDeleteKeyState({ open: false, tokenId: '' });
    } else {
      setDeleteKeyState(prev => ({ ...prev, open: value }));
    }
  };

  const handleRevokeClick = (tokenId: string) => {
    setDeleteKeyState({ open: true, tokenId });
  };

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex direction="column" className={sharedStyles.container}>
        <Flex direction="row" justify="between" align="center" className={sharedStyles.header}>
          <Flex gap={3} align="center">
            <Image
              alt="back-icon"
              style={{ cursor: 'pointer' }}
              src={backIcon as unknown as string}
              onClick={onBack}
              data-test-id="frontier-sdk-api-keys-page-back-link"
            />
            <PageHeader
              title="API"
            />
          </Flex>
        </Flex>
        <Flex justify="center" align="center">
          <Flex className={styles.content} direction="column" gap={9}>
          <Headings
            isLoading={isLoading}
            name={serviceUser?.title || ''}
            onManageAccessClick={() => setShowProjectsDialog(true)}
          />
          <AddServiceUserToken serviceUserId={serviceUserId} onAddToken={onAddToken} />
          <SerivceUserTokenList
            isLoading={isLoading}
            tokens={serviceUserTokens}
            onRevokeClick={handleRevokeClick}
          />
          </Flex>
        </Flex>
      </Flex>
      <ManageServiceUserProjectsDialog
        open={showProjectsDialog}
        onOpenChange={setShowProjectsDialog}
        serviceUserId={serviceUserId}
      />
      <DeleteServiceUserKeyDialog
        open={deleteKeyState.open}
        onOpenChange={handleDeleteKeyOpenChange}
        serviceUserId={serviceUserId}
        tokenId={deleteKeyState.tokenId}
      />
    </Flex>
  );
}
