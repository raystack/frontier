import { Button, Flex, Text, Skeleton, Image } from '@raystack/apsara/v1';
import styles from './styles.module.css';
import backIcon from '~/react/assets/chevron-left.svg';
import {
  Outlet,
  useLocation,
  useNavigate,
  useParams
} from '@tanstack/react-router';
import { FrontierClientAPIPlatformOptions } from '~/shared/types';
import { DEFAULT_API_PLATFORM_APP_NAME } from '~/react/utils/constants';
import { useCallback, useEffect, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1ServiceUser, V1Beta1ServiceUserToken } from '~/api-client/dist';
import AddServiceUserToken from './add-token';
import { CheckCircledIcon, CopyIcon } from '@radix-ui/react-icons';
import { useCopyToClipboard } from '~/react/hooks/useCopyToClipboard';

const Headings = ({
  isLoading,
  config,
  name,
  serviceUserId
}: {
  isLoading: boolean;
  name: string;
  config?: FrontierClientAPIPlatformOptions;
  serviceUserId: string;
}) => {
  const appName = config?.appName || DEFAULT_API_PLATFORM_APP_NAME;

  const navigate = useNavigate({ from: '/api-keys/$id' });

  function openProjectsPage() {
    navigate({
      to: '/api-keys/$id/projects',
      params: {
        id: serviceUserId
      }
    });
  }

  return (
    <Flex direction="column" gap="small" style={{ width: '100%' }}>
      {isLoading ? (
        <Skeleton containerClassName={styles.flex1} />
      ) : (
        <Flex justify={'between'}>
          <Text size="large">{name}</Text>
          <Button
            variant="outline"
            color="neutral"
            onClick={openProjectsPage}
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
          Create API key for accessing {appName} and its features
        </Text>
      )}
    </Flex>
  );
};

const ServiceUserTokenItem = ({
  token,
  isLoading,
  serviceUserId
}: {
  token: V1Beta1ServiceUserToken;
  isLoading: boolean;
  serviceUserId: string;
}) => {
  const [isCopied, setIsCopied] = useState(false);
  const { copy } = useCopyToClipboard();
  const navigate = useNavigate({ from: '/api-keys/$id' });

  function onRevokeClick() {
    navigate({
      to: '/api-keys/$id/key/$tokenId/delete',
      params: {
        tokenId: token?.id || '',
        id: serviceUserId
      }
    });
  }

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
    <Flex className={styles.serviceKeyItem} direction={'column'} gap={'small'}>
      <Flex justify={'between'} style={{ width: '100%' }} align={'center'}>
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
              onClick={onRevokeClick}
            >
              Revoke
            </Button>
          </>
        )}
      </Flex>
      {token?.token ? (
        <Flex gap={'small'} direction={'column'}>
          <Text size="small" variant="secondary" weight="regular">
            Note: Please save your key securely, it cannot be recovered after
            leaving this page
          </Text>
          <Flex className={styles.tokenBox} justify={'between'} gap={'medium'}>
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
  serviceUserId
}: {
  isLoading: boolean;
  tokens: V1Beta1ServiceUserToken[];
  serviceUserId: string;
}) => {
  const tokenList = isLoading
    ? [
        ...new Array(3).map(
          (_, i) => ({ id: i.toString() } as V1Beta1ServiceUserToken)
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
          serviceUserId={serviceUserId}
        />
      ))}
    </Flex>
  );
};

export default function ServiceUserPage() {
  let { id } = useParams({ from: '/api-keys/$id' });
  const { client, config, activeOrganization } = useFrontier();
  const navigate = useNavigate({ from: '/api-keys/$id' });

  const [serviceUser, setServiceUser] = useState<V1Beta1ServiceUser>();
  const [isServiceUserLoadning, setIsServiceUserLoading] = useState(false);

  const [serviceUserTokens, setServiceUserTokens] = useState<
    V1Beta1ServiceUserToken[]
  >([]);

  const [isServiceUserTokensLoading, setIsServiceUserTokensLoading] =
    useState(false);

  const location = useLocation();
  const existingToken = location?.state?.token;
  const refetch = location?.state?.refetch;
  const orgId = activeOrganization?.id || '';

  const getServiceUser = useCallback(
    async (serviceUserId: string) => {
      try {
        setIsServiceUserLoading(true);
        const resp = await client?.frontierServiceGetServiceUser(
          orgId,
          serviceUserId
        );
        const data = resp?.data?.serviceuser;
        setServiceUser(data);
      } catch (error) {
        console.error(error);
      } finally {
        setIsServiceUserLoading(false);
      }
    },
    [client, orgId]
  );

  const getServiceUserTokens = useCallback(
    async (serviceUserId: string) => {
      try {
        setIsServiceUserTokensLoading(true);
        const resp = await client?.frontierServiceListServiceUserTokens(
          orgId,
          serviceUserId
        );
        const data = resp?.data?.tokens || [];
        setServiceUserTokens(data);
      } catch (error) {
        console.error(error);
      } finally {
        setIsServiceUserTokensLoading(false);
      }
    },
    [client, orgId]
  );

  useEffect(() => {
    if (id) {
      getServiceUser(id);
      if (!existingToken?.id) {
        getServiceUserTokens(id);
      }
    }
  }, [id, getServiceUser, getServiceUserTokens, existingToken?.id, refetch]);

  const tokenList = existingToken
    ? [existingToken, ...serviceUserTokens]
    : serviceUserTokens;

  const isLoading = isServiceUserLoadning || isServiceUserTokensLoading;

  const onAddToken = (token: V1Beta1ServiceUserToken) => {
    setServiceUserTokens(prev => [token, ...prev]);
  };

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex className={styles.header} gap="small">
        <Image
          alt="back-icon"
          style={{ cursor: 'pointer' }}
          src={backIcon as unknown as string}
          onClick={() => navigate({ to: '/api-keys' })}
          data-test-id="frontier-sdk-api-keys-page-back-link"
        />
        <Text size="large">API</Text>
      </Flex>
      <Flex justify="center" align="center">
        <Flex className={styles.content} direction="column" gap="large">
          <Headings
            isLoading={isLoading}
            name={serviceUser?.title || ''}
            config={config?.apiPlatform}
            serviceUserId={id}
          />
          <AddServiceUserToken serviceUserId={id} onAddToken={onAddToken} />
          <SerivceUserTokenList
            isLoading={isLoading}
            tokens={tokenList}
            serviceUserId={id}
          />
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}
