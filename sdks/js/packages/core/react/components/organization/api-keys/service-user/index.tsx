import { Flex, Text } from '@raystack/apsara/v1';
import { Image } from '@raystack/apsara';
import styles from './styles.module.css';
import backIcon from '~/react/assets/chevron-left.svg';
import { useNavigate, useParams } from '@tanstack/react-router';
import Skeleton from 'react-loading-skeleton';
import { FrontierClientAPIPlatformOptions } from '~/shared/types';
import { DEFAULT_API_PLATFORM_APP_NAME } from '~/react/utils/constants';
import { useEffect, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1ServiceUser } from '~/api-client/dist';

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

export default function ServiceUserPage() {
  let { id } = useParams({ from: '/api-keys/$id' });
  const { client, config } = useFrontier();
  const navigate = useNavigate({ from: '/api-keys/$id' });
  const [serviceUser, setServiceUser] = useState<V1Beta1ServiceUser>();
  const [isServiceUserLoadning, setIsServiceUserLoading] = useState(false);

  useEffect(() => {
    async function getServiceUser(serviceUserId: string) {
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
    }
    if (id) {
      getServiceUser(id);
    }
  }, [id, client]);

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
            isLoading={isServiceUserLoadning}
            name={serviceUser?.title || ''}
            config={config?.apiPlatform}
          />
        </Flex>
      </Flex>
    </Flex>
  );
}
