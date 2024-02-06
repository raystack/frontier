import { Dialog, Flex, Text, Image, Separator, Button } from '@raystack/apsara';
import styles from '../../organization.module.css';
import { useNavigate, useParams } from '@tanstack/react-router';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useCallback, useEffect, useMemo, useState } from 'react';
import dayjs, { ManipulateType } from 'dayjs';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import { V1Beta1Plan } from '~/src';
import Skeleton from 'react-loading-skeleton';
import { getPlanChangeAction } from '~/react/utils';
import planStyles from '../plans.module.css';
import { usePlans } from '../hooks/usePlans';
import { toast } from 'sonner';

export default function ConfirmPlanChange() {
  const navigate = useNavigate({ from: '/plans/confirm-change/$planId' });
  const { planId } = useParams({ from: '/plans/confirm-change/$planId' });
  const {
    activePlan,
    isActivePlanLoading,
    config,
    client,
    fetchActiveSubsciption
  } = useFrontier();
  const [newPlan, setNewPlan] = useState<V1Beta1Plan>();
  const [isNewPlanLoading, setIsNewPlanLoading] = useState(false);

  const { changePlan, isLoading: isChangePlanLoading } = usePlans();

  const newPlanMetadata = newPlan?.metadata as Record<string, number>;
  const activePlanMetadata = activePlan?.metadata as Record<string, number>;

  const planAction = getPlanChangeAction(
    Number(newPlanMetadata?.weightage) || 0,
    Number(activePlanMetadata?.weightage)
  );

  const cancel = useCallback(() => navigate({ to: '/plans' }), [navigate]);

  const expiryDate = useMemo(() => {
    if (activePlan?.created_at && activePlan?.interval) {
      return dayjs(activePlan?.created_at)
        .add(1, activePlan?.interval as ManipulateType)
        .format(config.dateFormat || DEFAULT_DATE_FORMAT);
    }
    return '';
  }, [activePlan?.created_at, activePlan?.interval, config.dateFormat]);

  const verifyChange = useCallback(async () => {
    const activeSub = await fetchActiveSubsciption();
    const actionName = planAction?.btnLabel.toLowerCase();
    if (activeSub) {
      const planPhase = activeSub.phases?.find(
        phase => phase?.plan_id === planId
      );
      if (planPhase) {
        const changeDate = dayjs(planPhase?.effective_at).format(
          config?.dateFormat || DEFAULT_DATE_FORMAT
        );
        toast.success(`Plan ${actionName} successful`, {
          description: `Your plan will ${actionName} on ${changeDate}`
        });
        cancel();
      }
    }
  }, [
    cancel,
    config?.dateFormat,
    fetchActiveSubsciption,
    planAction?.btnLabel,
    planId
  ]);

  const onConfirm = useCallback(() => {
    changePlan({
      planId,
      onSuccess: verifyChange,
      immediate: planAction.immediate
    });
  }, [changePlan, planId, planAction.immediate, verifyChange]);

  const getPlan = useCallback(
    async (planId: string) => {
      setIsNewPlanLoading(true);

      try {
        const resp = await client?.frontierServiceGetPlan(planId);
        const plan = resp?.data?.plan;
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
    [client]
  );

  useEffect(() => {
    if (planId) {
      getPlan(planId);
    }
  }, [getPlan, planId]);

  const isLoading = isActivePlanLoading || isNewPlanLoading;

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
                {activePlan?.title}
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
                {newPlan?.title} (effective from {expiryDate})
              </Text>
            </Flex>
          )}
        </Flex>
        <Separator />
        <Flex justify={'end'} gap="medium" style={{ padding: 'var(--pd-16)' }}>
          <Button variant={'secondary'} onClick={cancel} size={'medium'}>
            Cancel
          </Button>
          <Button variant={'primary'} size={'medium'} onClick={onConfirm}>
            {isChangePlanLoading
              ? `${planAction?.btnLoadingLabel}...`
              : planAction?.btnLabel}
          </Button>
        </Flex>
      </Dialog.Content>
    </Dialog>
  );
}
