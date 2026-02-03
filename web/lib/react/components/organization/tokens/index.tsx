import { useEffect, useMemo } from 'react';
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
import { PageHeader } from '~/react/components/common/page-header';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { AuthTooltipMessage, getFormattedNumberString } from '~/react/utils';
import {
  FrontierServiceQueries,
  ListBillingTransactionsRequestSchema
} from '~/src';
import { TransactionsTable } from './transactions';
import { PlusIcon } from '@radix-ui/react-icons';
import { useBillingPermission } from '~/react/hooks/useBillingPermission';
import { useTokens } from '~/react/hooks/useTokens';
import { useNavigate, Outlet } from '@tanstack/react-router';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import coin from '~/react/assets/coin.svg';
import sharedStyles from '../styles.module.css';
import tokenStyles from './token.module.css';

interface TokenHeaderProps {
  billingSupportEmail?: string;
  isLoading?: boolean;
}

const TokensHeader = ({ billingSupportEmail, isLoading }: TokenHeaderProps) => {
  if (isLoading) {
    return (
      <Flex direction="column" gap={2} className={tokenStyles.flex1}>
        <Skeleton />
        <Skeleton />
      </Flex>
    );
  }

  return (
    <PageHeader
      title="Tokens"
      description={
        billingSupportEmail ? (
          <>
            Track your token balance, credit limit, and all transactions.{' '}
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
        ) : (
          'Track your token balance, credit limit, and all transactions.'
        )
      }
    />
  );
};

interface BalancePanelProps {
  balance: bigint;
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
  balance: bigint;
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
    data: transactionsRawData,
    isLoading: isTransactionsListLoading,
    error: transactionsError
  } = useQuery(
    FrontierServiceQueries.listBillingTransactions,
    create(ListBillingTransactionsRequestSchema, {
      orgId: activeOrganization?.id ?? '',
      expand: ['user']
    }),
    {
      enabled: !!activeOrganization?.id
    }
  );

  const transactionsData = useMemo(() => transactionsRawData?.transactions ?? [], [transactionsRawData]);

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
      <Flex direction="column" className={sharedStyles.container}>
        <Flex direction="row" justify="between" align="center" className={sharedStyles.header}>
          <TokensHeader
            billingSupportEmail={config.billing?.supportEmail}
            isLoading={isLoading}
          />
        </Flex>
        <Flex direction="column" gap={9}>
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
