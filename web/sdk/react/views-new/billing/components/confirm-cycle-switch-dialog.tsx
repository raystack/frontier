'use client';

import { useMemo } from 'react';
import {
  Button,
  Skeleton,
  Text,
  Flex,
  Dialog,
  toastManager
} from '@raystack/apsara-v1';
import { useFrontier } from '../../../contexts/FrontierContext';
import { getPlanIntervalName, getPlanPrice } from '../../../utils';
import { DEFAULT_DATE_FORMAT } from '../../../utils/constants';
import { timestampToDayjs } from '../../../../utils/timestamp';
import { usePlans } from '../../../views/plans/hooks/usePlans';
import { isEmpty } from 'lodash';

export interface ConfirmCycleSwitchPayload {
  planId: string;
}

type CycleSwitchDialogHandle = ReturnType<
  typeof Dialog.createHandle<ConfirmCycleSwitchPayload>
>;

export interface ConfirmCycleSwitchDialogProps {
  handle: CycleSwitchDialogHandle;
}

export function ConfirmCycleSwitchDialog({
  handle
}: ConfirmCycleSwitchDialogProps) {
  const {
    activePlan,
    paymentMethod,
    config,
    activeSubscription,
    allPlans,
    isAllPlansLoading
  } = useFrontier();
  const dateFormat = config?.dateFormat || DEFAULT_DATE_FORMAT;

  const {
    checkoutPlan,
    isLoading: isPlanActionLoading,
    changePlan,
    verifyPlanChange
  } = usePlans();

  const isLoading = isAllPlansLoading;

  return (
    <Dialog handle={handle}>
      {({ payload }) => {
        const planId = payload?.planId || '';

        return (
          <ConfirmCycleSwitchContent
            handle={handle}
            planId={planId}
            allPlans={allPlans}
            activePlan={activePlan}
            paymentMethod={paymentMethod}
            activeSubscription={activeSubscription}
            dateFormat={dateFormat}
            isLoading={isLoading}
            isPlanActionLoading={isPlanActionLoading}
            checkoutPlan={checkoutPlan}
            changePlan={changePlan}
            verifyPlanChange={verifyPlanChange}
          />
        );
      }}
    </Dialog>
  );
}

interface ConfirmCycleSwitchContentProps {
  handle: CycleSwitchDialogHandle;
  planId: string;
  allPlans: ReturnType<typeof useFrontier>['allPlans'];
  activePlan: ReturnType<typeof useFrontier>['activePlan'];
  paymentMethod: ReturnType<typeof useFrontier>['paymentMethod'];
  activeSubscription: ReturnType<typeof useFrontier>['activeSubscription'];
  dateFormat: string;
  isLoading: boolean;
  isPlanActionLoading: boolean;
  checkoutPlan: ReturnType<typeof usePlans>['checkoutPlan'];
  changePlan: ReturnType<typeof usePlans>['changePlan'];
  verifyPlanChange: ReturnType<typeof usePlans>['verifyPlanChange'];
}

function ConfirmCycleSwitchContent({
  handle,
  planId,
  allPlans,
  activePlan,
  paymentMethod,
  activeSubscription,
  dateFormat,
  isLoading,
  isPlanActionLoading,
  checkoutPlan,
  changePlan,
  verifyPlanChange
}: ConfirmCycleSwitchContentProps) {
  const nextPlan = useMemo(() => {
    if (planId && allPlans.length > 0) {
      return allPlans.find(p => p.id === planId);
    }
  }, [planId, allPlans]);

  const activePlanPrice = activePlan
    ? getPlanPrice(activePlan)
    : { amount: 0, currency: '' };
  const nextPlanPrice = nextPlan
    ? getPlanPrice(nextPlan)
    : { amount: 0, currency: '' };
  const isPaymentMethodRequired =
    isEmpty(paymentMethod) && nextPlanPrice.amount > 0;

  const activePlanIntervalName = getPlanIntervalName(activePlan);
  const nextPlanIntervalName = getPlanIntervalName(nextPlan);

  const savings = useMemo(() => {
    if (!activePlan || !nextPlan) return null;
    const activeInterval = activePlan.interval || '';
    const nextInterval = nextPlan.interval || '';
    if (activeInterval === nextInterval) return null;

    const toAnnual = (amount: number, interval: string) => {
      if (interval === 'month') return amount * 12;
      return amount;
    };
    const activeAnnual = toAnnual(
      Number(activePlanPrice.amount) || 0,
      activeInterval
    );
    const nextAnnual = toAnnual(
      Number(nextPlanPrice.amount) || 0,
      nextInterval
    );
    const diff = Math.abs(activeAnnual - nextAnnual);
    if (diff <= 0) return null;
    const cheaperInterval =
      activeAnnual < nextAnnual
        ? activePlanIntervalName
        : nextPlanIntervalName;
    const currency =
      activePlanPrice.currency || nextPlanPrice.currency || '';
    return { amount: diff, interval: cheaperInterval, currency };
  }, [
    activePlan,
    nextPlan,
    activePlanPrice,
    nextPlanPrice,
    activePlanIntervalName,
    nextPlanIntervalName
  ]);

  const nextPlanMetadata = nextPlan?.metadata as Record<string, number>;
  const activePlanMetadata = activePlan?.metadata as Record<string, number>;

  const isUpgrade =
    (Number(nextPlanMetadata?.weightage) || 0) -
    (Number(activePlanMetadata?.weightage) || 0) >
    0;

  const cycleSwitchDate = activeSubscription?.currentPeriodEndAt
    ? timestampToDayjs(activeSubscription?.currentPeriodEndAt)?.format(
      dateFormat
    )
    : 'the next billing cycle';

  async function onConfirm() {
    const nextPlanId = nextPlan?.id;
    if (!nextPlanId) return;
    if (isPaymentMethodRequired) {
      checkoutPlan({
        planId: nextPlanId,
        isTrial: false,
        onSuccess: data => {
          window.location.href = data?.checkoutUrl as string;
        }
      });
    } else {
      changePlan({
        planId: nextPlanId,
        onSuccess: async () => {
          const planPhase = await verifyPlanChange({
            planId: nextPlanId
          });
          if (planPhase) {
            handle.close();
            const changeDate = timestampToDayjs(
              planPhase?.effectiveAt
            )?.format(dateFormat);
            toastManager.add({
              title: 'Plan cycle switch successful',
              description: `Your plan cycle will be switched to ${nextPlanIntervalName} on ${changeDate}`,
              type: 'success'
            });
          }
        },
        immediate: isUpgrade
      });
    }
  }

  return (
    <Dialog.Content width={400}>
      <Dialog.Header>
        <Dialog.Title>Switch billing cycle</Dialog.Title>
      </Dialog.Header>
      <Dialog.Body>
        <Flex direction="column" gap={7}>
          {isLoading ? (
            <Skeleton />
          ) : (
            <Text size="small" variant="secondary">
              <Text as="span" size="small" weight="medium">
                Current cycle:
              </Text>{' '}
              {activePlanIntervalName}
            </Text>
          )}
          {isLoading ? (
            <Skeleton />
          ) : (
            <Text size="small" variant="secondary">
              <Text as="span" size="small" weight="medium">
                New cycle:
              </Text>{' '}
              {nextPlanIntervalName} (
              {isUpgrade
                ? 'effective immediately'
                : `effective from ${cycleSwitchDate}`}
              )
            </Text>
          )}
          {!isLoading && savings && (
            <Text size="small" variant="secondary">
              You can save{' '}
              <Text as="span" size="small" weight="medium">
                {savings.currency === 'usd' ? '$' : ''}
                {savings.amount}
              </Text>{' '}
              with {savings.interval} cycle.
            </Text>
          )}
        </Flex>
      </Dialog.Body>
      <Dialog.Footer>
        <Flex justify="end" gap={5}>
          <Button
            variant="outline"
            color="neutral"
            onClick={() => handle.close()}
            data-test-id="frontier-sdk-billing-cycle-switch-cancel-button"
          >
            Cancel
          </Button>
          <Button
            disabled={isLoading || isPlanActionLoading}
            onClick={onConfirm}
            loading={isPlanActionLoading}
            loaderText="Switching..."
            data-test-id="frontier-sdk-billing-cycle-switch-submit-button"
          >
            Switch cycle
          </Button>
        </Flex>
      </Dialog.Footer>
    </Dialog.Content>
  );
}
