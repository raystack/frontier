import {
  Button,
  Tooltip,
  Skeleton,
  Text,
  Headline,
  Flex,
  Image,
  toast,
  Link,
  Callout
} from '@raystack/apsara';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import { styles } from '../styles';
import tokenStyles from './token.module.css';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useEffect } from 'react';
import { AuthTooltipMessage, getFormattedNumberString } from '~/react/utils';
import {
  FrontierServiceQueries,
  ListBillingTransactionsRequestSchema
} from '~/src';
import { TransactionsTable } from './transactions';
import { PlusIcon } from '@radix-ui/react-icons';
import { useBillingPermission } from '~/react/hooks/useBillingPermission';
import { useTokens } from '~/react/hooks/useTokens';
import coin from '~/react/assets/coin.svg';
import { useNavigate, Outlet } from '@tanstack/react-router';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';

interface TokenHeaderProps {
  billingSupportEmail?: string;
  isLoading?: boolean;
}

const TokensHeader = ({ billingSupportEmail, isLoading }: TokenHeaderProps) => {
  return (
    <Flex direction="column" gap={3}>
      {isLoading ? (
        <Skeleton containerClassName={tokenStyles.flex1} />
      ) : (
        <Text size="large">Tokens</Text>
      )}
      {isLoading ? (
        <Skeleton containerClassName={tokenStyles.flex1} />
      ) : (
        <Text size="regular" variant="secondary">
          Track your token balance, credit limit, and all transactions.
          {billingSupportEmail ? (
            <>
              {' '}
              For more details, contact{' '}
              <Link
                size="regular"
                href={`mailto:${billingSupportEmail}`}
                data-test-id="frontier-sdk-billing-email-link"
                external
                style={{ textDecoration: 'none' }}
              >
                {billingSupportEmail}
              </Link>
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
  onAddTokenClick: () => void;
  canUpdateWorkspace: boolean;
}

function BalancePanel({
  balance,
  isLoading,
  onAddTokenClick,
  canUpdateWorkspace
}: BalancePanelProps) {
  const formattedBalance = getFormattedNumberString(balance);
  const disableAddTokensBtn = isLoading || !canUpdateWorkspace;
  return (
    <Flex className={tokenStyles.balancePanel} justify="between">
      <Flex className={tokenStyles.balanceTokenBox}>
        <Image
          src={coin as unknown as string}
          alt="coin"
          className={tokenStyles.coinIcon}
        />
        <Flex direction="column" gap={2}>
          <Text weight="medium" variant="secondary">
            Available tokens
          </Text>
          {isLoading ? (
            <Skeleton style={{ height: '24px' }} />
          ) : (
            <Headline size="t2">{formattedBalance}</Headline>
          )}
        </Flex>
      </Flex>
      <Flex>
        <Tooltip message={AuthTooltipMessage} disabled={canUpdateWorkspace}>
          {isLoading ? (
            <Skeleton height="28px" width="72px" />
          ) : (
            <Button
              variant="outline"
              color="neutral"
              size="small"
              className={tokenStyles.addTokenButton}
              onClick={onAddTokenClick}
              disabled={disableAddTokensBtn}
              loaderText="Adding tokens..."
              data-test-id="frontier-sdk-add-tokens-btn"
            >
              <Flex gap={2} align="center">
                <PlusIcon /> Add tokens
              </Flex>
            </Button>
          )}
        </Tooltip>
      </Flex>
    </Flex>
  );
}

interface TokenInfoBoxProps {
  balance: number;
  isLoading: boolean;
  canUpdateWorkspace: boolean;
}

function TokenInfoBox({ canUpdateWorkspace }: TokenInfoBoxProps) {
  const { billingDetails } = useFrontier();
  const isPostpaid =
    billingDetails?.creditMin && billingDetails.creditMin < BigInt(0);

  return isPostpaid && canUpdateWorkspace ? (
    <Callout
      type="accent"
      icon={<InfoCircledIcon className={tokenStyles.tokenInfoText} />}
      className={tokenStyles.tokenInfoBox}
    >
      You can now add tokens anytime to reduce next month’s invoice. But this
      won’t settle any existing or overdue invoices.
    </Callout>
  ) : null;
}

export default function Tokens() {
  const {
    config,
    activeOrganization,
    billingAccount,
    isActiveOrganizationLoading,
    isBillingAccountLoading
  } = useFrontier();
  const navigate = useNavigate({ from: '/tokens' });
  const { isAllowed, isFetching: isPermissionsFetching } =
    useBillingPermission();
  const { tokenBalance, isTokensLoading } = useTokens();

  const {
    data: transactionsData = [],
    isLoading: isTransactionsListLoading,
    error: transactionsError
  } = useQuery(
    FrontierServiceQueries.listBillingTransactions,
    create(ListBillingTransactionsRequestSchema, {
      orgId: activeOrganization?.id ?? '',
      expand: ['user'],
      billingId: billingAccount?.id ?? ''
    }),
    {
      enabled: !!activeOrganization?.id && !!billingAccount?.id,
      select: data => data?.transactions ?? []
    }
  );

  useEffect(() => {
    if (transactionsError) {
      console.error(transactionsError);
      toast.error('Unable to fetch transactions');
    }
  }, [transactionsError]);

  const isLoading =
    isActiveOrganizationLoading ||
    isBillingAccountLoading ||
    isTokensLoading ||
    isPermissionsFetching;

  const isTxnDataLoading = isLoading || isTransactionsListLoading;

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size="large">Tokens</Text>
      </Flex>
      <Flex direction="column" gap={9} style={styles.container}>
        <Flex direction="column" gap={9}>
          <TokensHeader
            billingSupportEmail={config.billing?.supportEmail}
            isLoading={isLoading}
          />
          <TokenInfoBox
            balance={tokenBalance}
            isLoading={isLoading}
            canUpdateWorkspace={isAllowed}
          />
          <BalancePanel
            balance={tokenBalance}
            isLoading={isLoading}
            onAddTokenClick={() => navigate({ to: '/tokens/modal' })}
            canUpdateWorkspace={isAllowed}
          />
          <TransactionsTable
            transactions={transactionsData}
            isLoading={isTxnDataLoading}
          />
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}
