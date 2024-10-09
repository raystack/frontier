import { useCallback, useEffect, useMemo, useState } from 'react';
import { Dialog, Flex, Text, Image, Separator, Button } from '@raystack/apsara';
import Skeleton from 'react-loading-skeleton';
import { useNavigate, useParams } from '@tanstack/react-router';
import * as _ from 'lodash';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import dayjs from 'dayjs';
import {
  DEFAULT_DATE_FORMAT,
  DEFAULT_PLAN_UPGRADE_MESSAGE
} from '~/react/utils/constants';
import { V1Beta1Plan } from '~/src';
import { getPlanChangeAction, getPlanNameWithInterval } from '~/react/utils';
import planStyles from '../plans.module.css';
import { usePlans } from '../hooks/usePlans';
import { toast } from 'sonner';
import styles from '../../organization.module.css';

export default function ConfirmPlanChange() {
  const navigate = useNavigate({ from: '/plans/confirm-change/$planId' });
  const { planId } = useParams({ from: '/plans/confirm-change/$planId' });
  const {
    activePlan,
    isActivePlanLoading,
    config,
    client,
    activeSubscription,
    basePlan
  } = useFrontier();
  const [newPlan, setNewPlan] = useState<V1Beta1Plan>();
  const [isNewPlanLoading, setIsNewPlanLoading] = useState(false);

  const {
    changePlan,
    isLoading: isChangePlanLoading,
    verifyPlanChange,
    verifySubscriptionCancel,
    cancelSubscription,
    checkBasePlan
  } = usePlans();

  const currentPlan = useMemo(() => allPlans.find((plan) => plan.id === planId), [ allPlans, planId ])

  const isNewPlanBasePlan = checkBasePlan(planId);

  const newPlanMetadata = newPlan?.metadata as Record<string, number>;
  const activePlanMetadata = activePlan?.metadata as Record<string, number>;

  const planAction = getPlanChangeAction(
    Number(newPlanMetadata?.weightage),
    Number(activePlanMetadata?.weightage)
  );

  const cancel = useCallback(() => navigate({ to: '/plans' }), [navigate]);

  const planChangeSlug =
    activePlan?.name && newPlan?.name
      ? `${activePlan?.name}:${newPlan?.name}`
      : '';

  const planChangeMessage = planChangeSlug
    ? _.get(config, ['messages', 'billing', 'plan_change', planChangeSlug])
    : '';

  const isUpgrade = planAction.btnLabel === 'Upgrade';

  const verifyChange = useCallback(async () => {
    const planPhase = isNewPlanBasePlan
      ? await verifySubscriptionCancel({})
      : await verifyPlanChange({ planId });
    const actionName = planAction?.btnLabel.toLowerCase();
    if (planPhase) {
      const changeDate = dayjs(planPhase?.effective_at).format(
        config?.dateFormat || DEFAULT_DATE_FORMAT
      );
      toast.success(`Plan ${actionName} successful`, {
        description: `Your plan will ${actionName} on ${changeDate}`
      });
      cancel();
    }
  }, [
    cancel,
    config?.dateFormat,
    planAction?.btnLabel,
    planId,
    verifyPlanChange,
    verifySubscriptionCancel,
    isNewPlanBasePlan
  ]);

  const onConfirm = useCallback(async () => {
    if (isNewPlanBasePlan) {
      cancelSubscription({
        onSuccess: verifyChange
      });
    } else {
      changePlan({
        planId,
        onSuccess: verifyChange,
        immediate: planAction.immediate
      });
    }
  }, [
    isNewPlanBasePlan,
    cancelSubscription,
    verifyChange,
    changePlan,
    planId,
    planAction.immediate
  ]);

  const getPlan = useCallback(
    () => {
      setIsNewPlanLoading(true);
      try {
        const plan = isNewPlanBasePlan
          ? basePlan
          : currentPlan
        if (plan) {
          setNewPlan(plan);
        }
      } catch (err) {
        console.error(
          'frontier:sdk:: There is problem with fetching active plan'
        );
        console.error(err);
      } finally {
        setIsNewPlanLoading(false);
      }
    },
    [isNewPlanBasePlan, basePlan, client]
  );

  useEffect(() => {
    if (planId) {
      getPlan();
    }
  }, [getPlan, planId]);

  const isLoading = isActivePlanLoading || isNewPlanLoading;

  const currentPlanName = getPlanNameWithInterval(activePlan, {
    hyphenSeperated: true
  });

  const upcomingPlanName = getPlanNameWithInterval(newPlan, {
    hyphenSeperated: true
  });

  const cycleSwitchDate = activeSubscription?.current_period_end_at
    ? dayjs(activeSubscription?.current_period_end_at).format(
        config?.dateFormat || DEFAULT_DATE_FORMAT
      )
    : 'the next billing cycle';

  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%', zIndex: '60' }}
        overlayClassname={styles.overlay}
      >
        <Flex justify="between" style={{ padding: '16px 24px' }}>
          {isLoading ? (
            <Skeleton containerClassName={planStyles.flex1} />
          ) : (
            <Text size={6} style={{ fontWeight: '500' }}>
              Verify {planAction?.btnLabel}
            </Text>
          )}

          <Image
            alt="cross"
            style={{ cursor: 'pointer' }}
            // @ts-ignore
            src={cross}
            onClick={cancel}
            data-test-id="frontier-sdk-confirm-plan-change-close-button"
          />
        </Flex>
        <Separator />
        <Flex
          style={{ padding: 'var(--pd-32) 24px', gap: '24px' }}
          direction={'column'}
        >
          {isLoading ? (
            <Skeleton />
          ) : (
            <Flex gap="small">
              <Text size={2} weight={500}>
                Current plan:
              </Text>
              <Text size={2} style={{ color: 'var(--foreground-muted)' }}>
                {currentPlanName}
              </Text>
            </Flex>
          )}
          {isLoading ? (
            <Skeleton />
          ) : (
            <Flex gap="small">
              <Text size={2} weight={500}>
                New plan:
              </Text>
              <Text size={2} style={{ color: 'var(--foreground-muted)' }}>
                {upcomingPlanName} (
                {planAction?.immediate
                  ? 'effective immediately'
                  : `effective from ${cycleSwitchDate}`}
                )
              </Text>
            </Flex>
          )}
          {isLoading ? (
            <Skeleton count={2} />
          ) : (
            <Text size={2} style={{ color: 'var(--foreground-muted)' }}>
              {planChangeMessage || (isUpgrade && DEFAULT_PLAN_UPGRADE_MESSAGE)}
            </Text>
          )}
        </Flex>

        <Separator />
        <Flex justify={'end'} gap="medium" style={{ padding: 'var(--pd-16)' }}>
          <Button
            variant={'secondary'}
            onClick={cancel}
            size={'medium'}
            data-test-id="frontier-sdk-confirm-plan-change-cancel-button"
          >
            Cancel
          </Button>
          <Button
            variant={'primary'}
            size={'medium'}
            onClick={onConfirm}
            disabled={isLoading || isChangePlanLoading}
            data-test-id="frontier-sdk-confirm-plan-change-submit-button"
          >
            {isChangePlanLoading
              ? `${planAction?.btnLoadingLabel}...`
              : planAction?.btnLabel}
          </Button>
        </Flex>
      </Dialog.Content>
    </Dialog>
  );
}
