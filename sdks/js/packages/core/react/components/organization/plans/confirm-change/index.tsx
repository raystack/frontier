import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button,
  toast,
  Skeleton,
  Image,
  Text,
  Flex,
  Dialog
} from '@raystack/apsara/v1';
import { useNavigate, useParams } from '@tanstack/react-router';
import * as _ from 'lodash';
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
import cross from '~/react/assets/cross.svg';
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
    basePlan,
    allPlans
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

  const getPlan = useCallback(() => {
    setIsNewPlanLoading(true);
    try {
      const plan = isNewPlanBasePlan ? basePlan : currentPlan;
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
  }, [isNewPlanBasePlan, basePlan, client]);

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
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%' }}
        overlayClassName={styles.overlay}
      >
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            {isLoading ? (
              <Skeleton containerClassName={planStyles.flex1} />
            ) : (
              <Text size="large" weight="medium">
                Verify {planAction?.btnLabel}
              </Text>
            )}

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              src={cross as unknown as string}
              onClick={cancel}
              data-test-id="frontier-sdk-confirm-plan-change-close-button"
            />
          </Flex>
        </Dialog.Header>

        <Dialog.Body>
          <Flex
            direction="column"
            gap={7}
          >
            {isLoading ? (
              <Skeleton />
            ) : (
              <Flex gap={3}>
                <Text size="small" weight="medium">
                  Current plan:
                </Text>
                <Text size="small" variant="secondary">
                  {currentPlanName}
                </Text>
              </Flex>
            )}
            {isLoading ? (
              <Skeleton />
            ) : (
              <Flex gap={3}>
                <Text size="small" weight="medium">
                  New plan:
                </Text>
                <Text size="small" variant="secondary">
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
              <Text size="small" variant="secondary">
                {planChangeMessage ||
                  (isUpgrade && DEFAULT_PLAN_UPGRADE_MESSAGE)}
              </Text>
            )}
          </Flex>
        </Dialog.Body>

        <Dialog.Footer>
          <Flex
            justify={'end'}
            gap={5}
          >
            <Button
              variant="outline"
              color="neutral"
              onClick={cancel}
              data-test-id="frontier-sdk-confirm-plan-change-cancel-button"
            >
              Cancel
            </Button>
            <Button
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
    </Dialog>
  );
}
