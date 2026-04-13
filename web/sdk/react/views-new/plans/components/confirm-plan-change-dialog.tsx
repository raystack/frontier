'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Amount,
  Button,
  Dialog,
  Flex,
  Skeleton,
  Text
} from '@raystack/apsara-v1';
import { toastManager } from '@raystack/apsara-v1';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { usePlans } from '~/react/views/plans/hooks/usePlans';
import { useMessages } from '~/react/hooks/useMessages';
import { getPlanChangeAction } from '~/react/utils';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import { timestampToDayjs } from '~/utils/timestamp';
import { Plan } from '@raystack/proton/frontier';

export interface ConfirmPlanChangePayload {
  planId: string;
  amount?: number;
  currency?: string;
}

export interface ConfirmPlanChangeDialogProps {
  handle: ReturnType<typeof Dialog.createHandle<ConfirmPlanChangePayload>>;
}

export function ConfirmPlanChangeDialog({
  handle
}: ConfirmPlanChangeDialogProps) {
  return (
    <Dialog handle={handle}>
      {({ payload }) => (
        <ConfirmPlanChangeContent
          planId={payload?.planId ?? ''}
          amount={payload?.amount}
          currency={payload?.currency}
          handle={handle}
        />
      )}
    </Dialog>
  );
}

interface ConfirmPlanChangeContentProps {
  planId: string;
  amount?: number;
  currency?: string;
  handle: ConfirmPlanChangeDialogProps['handle'];
}

function ConfirmPlanChangeContent({
  planId,
  amount,
  currency,
  handle
}: ConfirmPlanChangeContentProps) {
  const {
    activePlan,
    isAllPlansLoading,
    config,
    activeSubscription,
    paymentMethod,
    basePlan,
    allPlans
  } = useFrontier();

  const m = useMessages();

  const {
    changePlan,
    checkoutPlan,
    isLoading: isChangePlanLoading,
    verifyPlanChange,
    verifySubscriptionCancel,
    cancelSubscription,
    checkBasePlan
  } = usePlans();

  const [newPlan, setNewPlan] = useState<Plan>();
  const [isNewPlanLoading, setIsNewPlanLoading] = useState(false);

  const currentPlan = useMemo(
    () => allPlans.find(plan => plan.id === planId),
    [allPlans, planId]
  );

  const isNewPlanBasePlan = checkBasePlan(planId);
  const newPlanMetadata = newPlan?.metadata as Record<string, number>;
  const activePlanMetadata = activePlan?.metadata as Record<string, number>;

  const planAction = getPlanChangeAction(
    Number(newPlanMetadata?.weightage),
    Number(activePlanMetadata?.weightage)
  );

  const handleClose = useCallback(() => handle.close(), [handle]);

  const isUpgrade = planAction.btnLabel === 'Upgrade';

  const isAlreadySubscribed = activeSubscription?.planId !== undefined;
  const isCheckoutRequired =
    !paymentMethod ||
    (Object.keys(paymentMethod).length === 0 && (amount || 0) > 0);

  const newPlanSlug = isNewPlanBasePlan ? 'base' : newPlan?.name;
  const planChangeSlug = activePlan?.name
    ? `${activePlan?.name}:${newPlanSlug}`
    : '';
  const planChangeMessage = planChangeSlug
    ? m(`billing.plan_change.${planChangeSlug}`)
    : '';

  const defaultDowngradeMessage = newPlan?.title
    ? `Downgrading your plan to ${newPlan.title} will limit access.`
    : 'Downgrading your plan will limit access.';

  const verifyChange = useCallback(async () => {
    const planPhase = isNewPlanBasePlan
      ? await verifySubscriptionCancel({})
      : await verifyPlanChange({ planId });
    const actionName = planAction?.btnLabel.toLowerCase();
    if (planPhase) {
      const changeDate = timestampToDayjs(planPhase?.effectiveAt)?.format(
        config?.dateFormat || DEFAULT_DATE_FORMAT
      );
      toastManager.add({
        title: `Plan ${actionName} successful`,
        description: `Your plan will ${actionName} on ${changeDate}`,
        type: 'success'
      });
      handleClose();
    }
  }, [
    handleClose,
    config?.dateFormat,
    planAction?.btnLabel,
    planId,
    verifyPlanChange,
    verifySubscriptionCancel,
    isNewPlanBasePlan
  ]);

  const onConfirm = useCallback(async () => {
    if (isNewPlanBasePlan) {
      cancelSubscription({ onSuccess: verifyChange });
    } else if (isAlreadySubscribed && !isCheckoutRequired) {
      changePlan({
        planId,
        onSuccess: verifyChange,
        immediate: planAction.immediate
      });
    } else {
      checkoutPlan({
        planId,
        isTrial: false,
        onSuccess: data => {
          window.location.href = data?.checkoutUrl as string;
        }
      });
    }
  }, [
    isNewPlanBasePlan,
    isAlreadySubscribed,
    isCheckoutRequired,
    cancelSubscription,
    verifyChange,
    changePlan,
    checkoutPlan,
    planId,
    planAction.immediate
  ]);

  useEffect(() => {
    if (planId) {
      setIsNewPlanLoading(true);
      try {
        const plan = isNewPlanBasePlan ? basePlan : currentPlan;
        if (plan) setNewPlan(plan);
      } finally {
        setIsNewPlanLoading(false);
      }
    }
  }, [planId, isNewPlanBasePlan, basePlan, currentPlan]);

  const isLoading = isAllPlansLoading || isNewPlanLoading;

  const cycleSwitchDate = activeSubscription?.currentPeriodEndAt
    ? timestampToDayjs(activeSubscription?.currentPeriodEndAt)?.format(
        config?.dateFormat || DEFAULT_DATE_FORMAT
      )
    : 'the next billing cycle';

  const effectiveDateLabel = planAction?.immediate
    ? 'effective immediately'
    : `effective from ${cycleSwitchDate}`;

  return (
    <Dialog.Content width={400}>
      <Dialog.Header>
        <Dialog.Title>
          {isLoading ? (
            <Skeleton height="20px" width="150px" />
          ) : (
            `Verify ${planAction?.btnLabel}`
          )}
        </Dialog.Title>
      </Dialog.Header>

      <Dialog.Body>
        <Flex direction="column" gap={7}>
          {isLoading ? (
            <Skeleton height="16px" />
          ) : (
            <Text size="small">
              <Text size="small" weight="medium" as="span">
                Current plan:
              </Text>{' '}
              <Text size="small" variant="secondary" as="span">
                {activePlan?.title}
              </Text>
            </Text>
          )}

          {isLoading ? (
            <Skeleton height="16px" />
          ) : (
            <Text size="small">
              <Text size="small" weight="medium" as="span">
                New plan:
              </Text>{' '}
              <Text size="small" variant="secondary" as="span">
                {newPlan?.title} ({effectiveDateLabel})
              </Text>
            </Text>
          )}

          {isLoading ? (
            <Skeleton height="16px" />
          ) : isUpgrade ? (
            <Text size="small">
              <Text size="small" weight="medium" as="span">
                Price:
              </Text>{' '}
              <Text size="small" variant="secondary" as="span">
                <Amount
                  value={amount || 0}
                  currency={currency}
                  hideDecimals={config?.billing?.hideDecimals}
                  valueInMinorUnits={false}
                />
              </Text>
            </Text>
          ) : (
            <Text size="small" variant="secondary">
              {planChangeMessage || defaultDowngradeMessage}
            </Text>
          )}
        </Flex>
      </Dialog.Body>

      <Dialog.Footer>
        <Flex justify="end" gap={5}>
          <Button
            variant="outline"
            color="neutral"
            onClick={handleClose}
            data-test-id="frontier-sdk-confirm-plan-change-cancel-button"
          >
            Cancel
          </Button>
          <Button
            variant="solid"
            color="accent"
            onClick={onConfirm}
            disabled={isLoading || isChangePlanLoading}
            loading={isChangePlanLoading}
            loaderText={`${planAction?.btnLoadingLabel}...`}
            data-test-id="frontier-sdk-confirm-plan-change-submit-button"
          >
            {planAction?.btnLabel}
          </Button>
        </Flex>
      </Dialog.Footer>
    </Dialog.Content>
  );
}
