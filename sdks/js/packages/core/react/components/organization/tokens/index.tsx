import { Button, Tooltip, Skeleton, Text, Headline, Flex, Image, toast, Link, Callout } from '@raystack/apsara/v1';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import { styles } from '../styles';
import tokenStyles from './token.module.css';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useEffect, useState } from 'react';
import { AuthTooltipMessage, getFormattedNumberString } from '~/react/utils';
import { V1Beta1BillingTransaction } from '~/src';
import { TransactionsTable } from './transactions';
import { PlusIcon } from '@radix-ui/react-icons';
import qs from 'query-string';
import { DEFAULT_TOKEN_PRODUCT_NAME } from '~/react/utils/constants';
import { useBillingPermission } from '~/react/hooks/useBillingPermission';
import { useTokens } from '~/react/hooks/useTokens';
import coin from '~/react/assets/coin.svg';
import billingStyles from '../billing/billing.module.css';

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
          Oversee your billing and invoices.
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
  isCheckoutLoading: boolean;
  onAddTokenClick: () => void;
  canUpdateWorkspace: boolean;
}

function BalancePanel({
  balance,
  isLoading,
  isCheckoutLoading,
  onAddTokenClick,
  canUpdateWorkspace
}: BalancePanelProps) {
  const formattedBalance = getFormattedNumberString(balance);
  const disableAddTokensBtn =
    isLoading || isCheckoutLoading || !canUpdateWorkspace;
  return (
    <Flex className={tokenStyles.balancePanel} justify="between">
      <Flex className={tokenStyles.balanceTokenBox}>
        <Image src={coin as unknown as string} alt="coin" className={tokenStyles.coinIcon} />
        <Flex direction="column" gap={2}>
          <Text weight="medium" variant="secondary">
            Available tokens
          </Text>
          {isLoading ? (
            <Skeleton style={{ height: '24px' }} />
          ) : (
            <Headline size="t2" weight="medium">
              {formattedBalance}
            </Headline>
          )}
        </Flex>
      </Flex>
      <Flex>
        <Tooltip message={AuthTooltipMessage} disabled={canUpdateWorkspace}>
          {isLoading ? (
            <Skeleton height='28px' width='72px' />
          ) : (
          <Button
            variant="outline"
            color="neutral"
            size="small"
            className={tokenStyles.addTokenButton}
            onClick={onAddTokenClick}
            disabled={disableAddTokensBtn}
            loading={isCheckoutLoading}
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
  isCheckoutLoading: boolean;
  canUpdateWorkspace: boolean;
}

function TokenInfoBox({}: TokenInfoBoxProps) {
  const { billingDetails } = useFrontier();
  const isPostpaid = billingDetails?.credit_min && parseInt(billingDetails.credit_min) < 0
  return (
    <>
      {isPostpaid && (
      <Callout
        type="accent"
        icon={<InfoCircledIcon className={tokenStyles.tokenInfoText} />}
        className={tokenStyles.tokenInfoBox}
      >You can now add tokens anytime to reduce next month’s invoice. But this won’t settle any existing or overdue invoices.
      </Callout>)}
    </>
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
  const [transactionsList, setTransactionsList] = useState<
    V1Beta1BillingTransaction[]
  >([]);
  const [isTransactionsListLoading, setIsTransactionsListLoading] =
    useState(false);
  const [isCheckoutLoading, setIsCheckoutLoading] = useState(false);
  const { isAllowed, isFetching: isPermissionsFetching } =
    useBillingPermission();
  const { tokenBalance, isTokensLoading } = useTokens();

  useEffect(() => {
    async function getTransactions(orgId: string, billingAccountId: string) {
      try {
        setIsTransactionsListLoading(true);
        const resp = await client?.frontierServiceListBillingTransactions(
          orgId,
          billingAccountId,
          {
            expand: ['user']
          }
        );
        const txns = resp?.data?.transactions || [];
        setTransactionsList(txns);
      } catch (err: any) {
        console.error(err);
        toast.error('Unable to fetch transactions');
      } finally {
        setIsTransactionsListLoading(false);
      }
    }

    if (activeOrganization?.id && billingAccount?.id) {
      getTransactions(activeOrganization?.id, billingAccount?.id);
    }
  }, [activeOrganization?.id, billingAccount?.id, client]);

  const onAddTokenClick = async () => {
    setIsCheckoutLoading(true);
    try {
      if (activeOrganization?.id && billingAccount?.id) {
        // Token product id or name can be used here
        const tokenProductId =
          config?.billing?.tokenProductId || DEFAULT_TOKEN_PRODUCT_NAME;
        const query = qs.stringify(
          {
            details: btoa(
              qs.stringify({
                billing_id: billingAccount?.id,
                organization_id: activeOrganization?.id,
                type: 'tokens'
              })
            ),
            checkout_id: '{{.CheckoutID}}'
          },
          { encode: false }
        );
        const cancel_url = `${config?.billing?.cancelUrl}?${query}`;
        const success_url = `${config?.billing?.successUrl}?${query}`;

        const resp = await client?.frontierServiceCreateCheckout(
          activeOrganization?.id,
          billingAccount?.id,
          {
            cancel_url: cancel_url,
            success_url: success_url,
            product_body: {
              product: tokenProductId
            }
          }
        );
        if (resp?.data?.checkout_session?.checkout_url) {
          window.location.href = resp?.data?.checkout_session.checkout_url;
        }
      }
    } catch (err: any) {
      console.error(err);
      toast.error('Something went wrong', {
        description: err?.message
      });
    } finally {
      setIsCheckoutLoading(false);
    }
  };

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
            isCheckoutLoading={isCheckoutLoading}
            canUpdateWorkspace={isAllowed}
          />
          <BalancePanel
            balance={tokenBalance}
            isLoading={isLoading}
            onAddTokenClick={onAddTokenClick}
            isCheckoutLoading={isCheckoutLoading}
            canUpdateWorkspace={isAllowed}
          />
          <TransactionsTable
            transactions={transactionsList}
            isLoading={isTxnDataLoading}
          />
        </Flex>
      </Flex>
    </Flex>
  );
}
