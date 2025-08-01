import { useCallback, useEffect, useState } from 'react';
import {
  Button,
  Skeleton,
  Image,
  Text,
  toast,
  Flex,
  Dialog
} from '@raystack/apsara/v1';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Plan } from '~/src';
import { getPlanIntervalName, getPlanPrice } from '~/react/utils';
import * as _ from 'lodash';
import { usePlans } from '../../plans/hooks/usePlans';
import dayjs from 'dayjs';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import cross from '~/react/assets/cross.svg';
import styles from '../../organization.module.css';

export function ConfirmCycleSwitch() {
  const { activePlan, client, paymentMethod, config, activeSubscription } =
    useFrontier();
  const navigate = useNavigate({ from: '/billing/cycle-switch/$planId' });
  const { planId } = useParams({ from: '/billing/cycle-switch/$planId' });
  const dateFormat = config?.dateFormat || DEFAULT_DATE_FORMAT;

  const [isPlanLoading, setIsPlanLoading] = useState(false);
  const [nextPlan, setNextPlan] = useState<V1Beta1Plan>();
  const [isCycleSwitching, setCycleSwitching] = useState(false);

  const closeModal = useCallback(
    () => navigate({ to: '/billing' }),
    [navigate]
  );

  const {
    checkoutPlan,
    isLoading: isPlanActionLoading,
    changePlan,
    verifyPlanChange
  } = usePlans();

  const nextPlanPrice = nextPlan ? getPlanPrice(nextPlan) : { amount: 0 };
  const isPaymentMethodRequired =
    _.isEmpty(paymentMethod) && nextPlanPrice.amount > 0;

  const nextPlanIntervalName = getPlanIntervalName(nextPlan);

  const nextPlanMetadata = nextPlan?.metadata as Record<string, number>;
  const activePlanMetadata = activePlan?.metadata as Record<string, number>;

  const isUpgrade =
    (Number(nextPlanMetadata?.weightage) || 0) -
      (Number(activePlanMetadata?.weightage) || 0) >
    0;

  useEffect(() => {
    async function getNextPlan(nextPlanId: string) {
      setIsPlanLoading(true);
      try {
        const resp = await client?.frontierServiceGetPlan(nextPlanId);
        const plan = resp?.data?.plan;
        setNextPlan(plan);
      } catch (err: any) {
        toast.error('Something went wrong', {
          description: err.message
        });
        console.error(err);
      } finally {
        setIsPlanLoading(false);
      }
    }
    if (planId) {
      getNextPlan(planId);
    }
  }, [client, planId]);

  const isLoading = isPlanLoading;

  async function onConfirm() {
    setCycleSwitching(true);
    try {
      if (nextPlan?.id) {
        const nextPlanId = nextPlan?.id;
        if (isPaymentMethodRequired) {
          checkoutPlan({
            planId: nextPlanId,
            isTrial: false,
            onSuccess: data => {
              window.location.href = data?.checkout_url as string;
            }
          });
        } else
          changePlan({
            planId: nextPlanId,
            onSuccess: async () => {
              const planPhase = await verifyPlanChange({
                planId: nextPlanId
              });
              if (planPhase) {
                closeModal();
                const changeDate = dayjs(planPhase?.effective_at).format(
                  dateFormat
                );
                toast.success(`Plan cycle switch successful`, {
                  description: `Your plan cycle will switched to ${nextPlanIntervalName} on ${changeDate}`
                });
              }
            },
            immediate: isUpgrade
          });
      }
    } catch (err: any) {
      console.error(err);
      toast.error('Something went wrong', {
        description: err.message
      });
    } finally {
      setCycleSwitching(false);
    }
  }

  const cycleSwitchDate = activeSubscription?.current_period_end_at
    ? dayjs(activeSubscription?.current_period_end_at).format(
        config?.dateFormat || DEFAULT_DATE_FORMAT
      )
    : 'the next billing cycle';

  return (
    <Dialog open={true}>
      <Dialog.Content
        overlayClassName={styles.overlay}
        style={{ padding: 0, maxWidth: '600px', width: '100%' }}
      >
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="large" weight="medium">
              Switch billing cycle
            </Text>

            <Image
              data-test-id="frontier-sdk-billing-cycle-switch-close-button"
              alt="cross"
              style={{ cursor: 'pointer' }}
              src={cross as unknown as string}
              onClick={closeModal}
            />
          </Flex>
        </Dialog.Header>

        <Dialog.Body>
          <Flex
            direction={'column'}
            gap={7}
          >
            {isLoading ? (
              <Skeleton />
            ) : (
              <Flex gap={3}>
                <Text size="small" weight="medium">
                  Current cycle:
                </Text>
                <Text size="small" variant="secondary">
                  {getPlanIntervalName(activePlan)}
                </Text>
              </Flex>
            )}
            {isLoading ? (
              <Skeleton />
            ) : (
              <Flex gap={3}>
                <Text size="small" weight="medium">
                  New cycle:
                </Text>
                <Text size="small" variant="secondary">
                  {nextPlanIntervalName} (
                  {isUpgrade
                    ? 'effective immediately'
                    : `effective from ${cycleSwitchDate}`}
                  )
                </Text>
              </Flex>
            )}
          </Flex>
        </Dialog.Body>

        <Dialog.Footer>
          <Flex
            justify="end"
            gap={5}
          >
            <Button
              variant="outline"
              color="neutral"
              onClick={closeModal}
              data-test-id="frontier-sdk-billing-cycle-switch-cancel-button"
            >
              Cancel
            </Button>
            <Button
              disabled={isLoading || isCycleSwitching || isPlanActionLoading}
              onClick={onConfirm}
              loading={isCycleSwitching}
              loaderText="Switching..."
              data-test-id="frontier-sdk-billing-cycle-switch-submit-button"
            >
              Switch cycle
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
}
