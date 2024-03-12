import { Flex, Image, Text } from '@raystack/apsara';
import { styles } from '../styles';
import Skeleton from 'react-loading-skeleton';
import tokenStyles from './token.module.css';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useEffect, useState } from 'react';
import coin from '~/react/assets/coin.svg';
import { getFormattedNumberString } from '~/react/utils';

interface TokenHeaderProps {
  billingSupportEmail?: string;
  isLoading?: boolean;
}

const TokensHeader = ({ billingSupportEmail, isLoading }: TokenHeaderProps) => {
  return (
    <Flex direction="column" gap="small">
      {isLoading ? (
        <Skeleton containerClassName={tokenStyles.flex1} />
      ) : (
        <Text size={6}>Tokens</Text>
      )}
      {isLoading ? (
        <Skeleton containerClassName={tokenStyles.flex1} />
      ) : (
        <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
          Oversee your billing and invoices.
          {billingSupportEmail ? (
            <>
              {' '}
              For more details, contact{' '}
              <a
                href={`mailto:${billingSupportEmail}`}
                target="_blank"
                style={{ fontWeight: 400, color: 'var(--foreground-accent)' }}
              >
                {billingSupportEmail}
              </a>
            </>
          ) : null}
        </Text>
      )}
    </Flex>
  );
};

interface BalancePanelProps {
  balance: number;
  isLoading: boolean;
}

function BalancePanel({ balance, isLoading }: BalancePanelProps) {
  const formattedBalance = getFormattedNumberString(balance);
  return (
    <Flex className={tokenStyles.balancePanel}>
      <Flex className={tokenStyles.balanceTokenBox}>
        {/* @ts-ignore */}
        <Image src={coin} alt="coin" className={tokenStyles.coinIcon} />
        <Flex direction={'column'} gap={'extra-small'}>
          <Text weight={500} style={{ color: 'var(--foreground-muted)' }}>
            Available tokens
          </Text>
          {isLoading ? (
            <Skeleton style={{ height: '24px' }} />
          ) : (
            <Text size={9} weight={600}>
              {formattedBalance}
            </Text>
          )}
        </Flex>
      </Flex>
    </Flex>
  );
}

export default function Tokens() {
  const {
    config,
    client,
    activeOrganization,
    billingAccount,
    isActiveOrganizationLoading,
    isBillingAccountLoading
  } = useFrontier();
  const [tokenBalance, setTokenBalance] = useState(0);
  const [isTokensLoading, setIsTokensLoading] = useState(false);

  useEffect(() => {
    async function getBalance(orgId: string, billingAccountId: string) {
      try {
        setIsTokensLoading(true);
        const resp = await client?.frontierServiceGetBillingBalance(
          orgId,
          billingAccountId
        );
        const tokens = resp?.data?.balance?.amount || '0';
        setTokenBalance(Number(tokens));
      } catch (err: any) {
        console.error(err);
      } finally {
        setIsTokensLoading(false);
      }
    }

    if (activeOrganization?.id && billingAccount?.id) {
      getBalance(activeOrganization?.id, billingAccount?.id);
    }
  }, [activeOrganization?.id, billingAccount?.id, client]);

  const isLoading =
    isActiveOrganizationLoading || isBillingAccountLoading || isTokensLoading;

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Tokens</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <TokensHeader
            billingSupportEmail={config.billing?.supportEmail}
            isLoading={isLoading}
          />
          <BalancePanel balance={tokenBalance} isLoading={isLoading} />
        </Flex>
      </Flex>
    </Flex>
  );
}
