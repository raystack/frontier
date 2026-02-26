import { useCallback, useMemo } from 'react';
import {
  Button,
  Skeleton,
  Image,
  Text,
  toast,
  Flex,
  Dialog
} from '@raystack/apsara';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { getPlanIntervalName, getPlanPrice } from '~/react/utils';
import * as _ from 'lodash';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import cross from '~/react/assets/cross.svg';
import orgStyles from '../../components/organization/organization.module.css';
import { timestampToDayjs } from '~/utils/timestamp';
import { usePlans } from '~/react/views/plans/hooks/usePlans';

export interface ConfirmCycleSwitchDialogProps {
  open: boolean;
  onOpenChange?: (value: boolean) => void;
  planId: string;
}

export function ConfirmCycleSwitchDialog({
  open,
  onOpenChange,
  planId
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

  const handleClose = useCallback(
    () => onOpenChange?.(false),
    [onOpenChange]
  );

  const {
    checkoutPlan,
    isLoading: isPlanActionLoading,
    changePlan,
    verifyPlanChange
  } = usePlans();

  const nextPlan = useMemo(() => {
    if (planId && allPlans.length > 0) {
      const plan = allPlans.find(p => p.id === planId);
      return plan;
    }
  }, [planId, allPlans]);

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

  const isLoading = isAllPlansLoading;

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
    } else
      changePlan({
        planId: nextPlanId,
        onSuccess: async () => {
          const planPhase = await verifyPlanChange({
            planId: nextPlanId
          });
          if (planPhase) {
            handleClose();
            const changeDate = timestampToDayjs(planPhase?.effectiveAt)?.format(
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

  const cycleSwitchDate = activeSubscription?.currentPeriodEndAt
    ? timestampToDayjs(activeSubscription?.currentPeriodEndAt)?.format(
      config?.dateFormat || DEFAULT_DATE_FORMAT
    )
    : 'the next billing cycle';

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <Dialog.Content
        overlayClassName={orgStyles.overlay}
        style={{ padding: 0, maxWidth: '600px', width: '100%' }}
      >
        <Dialog.Header>
          <Flex justify="between" align="center" width="full">
            <Text size="large" weight="medium">
              Switch billing cycle
            </Text>

            <Image
              data-test-id="frontier-sdk-billing-cycle-switch-close-button"
              alt="cross"
              style={{ cursor: 'pointer' }}
              src={cross as unknown as string}
              onClick={handleClose}
            />
          </Flex>
        </Dialog.Header>

        <Dialog.Body>
          <Flex direction={'column'} gap={7}>
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
          <Flex justify="end" gap={5}>
            <Button
              variant="outline"
              color="neutral"
              onClick={handleClose}
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
    </Dialog>
  );
}
