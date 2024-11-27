import { Button, Flex, Text } from '@raystack/apsara/v1';
import { Image, TextField } from '@raystack/apsara';
import styles from './styles.module.css';
import backIcon from '~/react/assets/chevron-left.svg';
import { Outlet, useNavigate, useParams } from '@tanstack/react-router';
import Skeleton from 'react-loading-skeleton';
import { FrontierClientAPIPlatformOptions } from '~/shared/types';
import { DEFAULT_API_PLATFORM_APP_NAME } from '~/react/utils/constants';
import { useCallback, useEffect, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1ServiceUser, V1Beta1ServiceUserToken } from '~/api-client/dist';

const Headings = ({
  isLoading,
  config,
  name
}: {
  isLoading: boolean;
  name: string;
  config?: FrontierClientAPIPlatformOptions;
}) => {
  const appName = config?.appName || DEFAULT_API_PLATFORM_APP_NAME;
  return (
    <Flex direction="column" gap="small" style={{ width: '100%' }}>
      {isLoading ? (
        <Skeleton containerClassName={styles.flex1} />
      ) : (
        <Text size={6}>{name}</Text>
      )}
      {isLoading ? (
        <Skeleton containerClassName={styles.flex1} />
      ) : (
        <Text size={4} variant="secondary">
          Create API key for accessing {appName} and its features
        </Text>
      )}
    </Flex>
  );
};

const SerivceUserTokenList = ({
  isLoading,
  tokens
}: {
  isLoading: boolean;
  tokens: V1Beta1ServiceUserToken[];
}) => {
  const tokenList = isLoading
    ? [
        ...new Array(3).map(
          (_, i) => ({ id: i.toString() } as V1Beta1ServiceUserToken)
        )
      ]
    : tokens;

  return (
    <Flex direction="column">
      {tokenList.map((token, i) => (
        <Flex
          key={token?.id}
          className={styles.serviceKeyItem}
          direction={'column'}
          gap={'small'}
        >
          <Flex justify={'between'} style={{ width: '100%' }} align={'center'}>
            {isLoading ? (
              <>
                <Skeleton containerClassName={styles.flex1} width={'300px'} />
                <Skeleton containerClassName={styles.serviceKeyItemLoaderBtn} />
              </>
            ) : (
              <>
                <Text size={3} weight={500}>
                  {token?.title}
                </Text>
                <Button
                  variant="secondary"
                  size={'small'}
                  data-test-id={`frontier-sdk-service-account-token-revoke-btn-${i}`}
                >
                  Revoke
                </Button>
              </>
            )}
          </Flex>
          {token?.token ? (
            <Flex gap={'small'} direction={'column'}>
              <Text size={2} variant={'secondary'} weight={400}>
                Note: Please save your key securely, it cannot be recovered
                after leaving this page
              </Text>
              <Text size={3} weight={500}>
                {token?.token}
              </Text>
            </Flex>
          ) : null}
        </Flex>
      ))}
    </Flex>
  );
};

export default function ServiceUserPage() {
  let { id } = useParams({ from: '/api-keys/$id' });
  const { client, config } = useFrontier();
  const navigate = useNavigate({ from: '/api-keys/$id' });
  const [serviceUser, setServiceUser] = useState<V1Beta1ServiceUser>();
  const [isServiceUserLoadning, setIsServiceUserLoading] = useState(false);

  const [serviceUserTokens, setServiceUserTokens] = useState<
    V1Beta1ServiceUserToken[]
  >([]);

  const [isServiceUserTokensLoading, setIsServiceUserTokensLoading] =
    useState(false);

  const getServiceUser = useCallback(
    async (serviceUserId: string) => {
      try {
        setIsServiceUserLoading(true);
        const resp = await client?.frontierServiceGetServiceUser(serviceUserId);
        const data = resp?.data?.serviceuser;
        setServiceUser(data);
      } catch (error) {
        console.error(error);
      } finally {
        setIsServiceUserLoading(false);
      }
    },
    [client]
  );

  const getServiceUserTokens = useCallback(
    async (serviceUserId: string) => {
      try {
        setIsServiceUserTokensLoading(true);
        const resp = await client?.frontierServiceListServiceUserTokens(
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
    [client]
  );

  useEffect(() => {
    if (id) {
      getServiceUser(id);
      getServiceUserTokens(id);
    }
  }, [id, getServiceUser, getServiceUserTokens]);

  const isLoading = isServiceUserLoadning || isServiceUserTokensLoading;

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex className={styles.header} gap="small">
        <Image
          alt="back-icon"
          style={{ cursor: 'pointer' }}
          // @ts-ignore
          src={backIcon}
          onClick={() => navigate({ to: '/api-keys' })}
          data-test-id="frontier-sdk-api-keys-page-back-link"
        />
        <Text size={6}>API</Text>
      </Flex>
      <Flex justify="center" align="center">
        <Flex className={styles.content} direction="column" gap="large">
          <Headings
            isLoading={isLoading}
            name={serviceUser?.title || ''}
            config={config?.apiPlatform}
          />
          <Flex gap="small">
            <TextField
              size={'medium'}
              placeholder="Search key name"
            ></TextField>
            <Button
              data-test-id="frontier-sdk-api-keys-new-token-btn"
              onClick={() =>
                navigate({
                  to: '/api-keys/$id/add-token'
                })
              }
            >
              Generate new key
            </Button>
          </Flex>
          <SerivceUserTokenList
            isLoading={isLoading}
            tokens={serviceUserTokens}
          />
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}
