import { Dialog, Flex, Text, Image, Separator, Button } from '@raystack/apsara';
import Skeleton from 'react-loading-skeleton';
import { useNavigate, useParams } from '@tanstack/react-router';
import cross from '~/react/assets/cross.svg';
import styles from '../../organization.module.css';
import { useCallback, useEffect, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Plan } from '~/src';
import { toast } from 'sonner';
import { getPlanIntervalName } from '~/react/utils';

export function ConfirmCycleSwitch() {
  const { activePlan, client } = useFrontier();

  const navigate = useNavigate({ from: '/billing/cycle-switch/$planId' });
  const { planId } = useParams({ from: '/billing/cycle-switch/$planId' });
  const cancel = useCallback(() => navigate({ to: '/billing' }), [navigate]);

  const [isPlanLoading, setIsPlanLoading] = useState(false);
  const [nextPlan, setNextPlan] = useState<V1Beta1Plan>();

  const [isCycleSwitching, setCycleSwitching] = useState(false);

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
    } catch (err: any) {
      console.error(err);
      toast.error('Something went wrong', {
        description: err.message
      });
    } finally {
      setCycleSwitching(false);
    }
  }

  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%', zIndex: '60' }}
        overlayClassname={styles.overlay}
      >
        <Flex justify="between" style={{ padding: '16px 24px' }}>
          <Text size={6} style={{ fontWeight: '500' }}>
            Switch billing cycle
          </Text>

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
                Current cycle:
              </Text>
              <Text size={2} style={{ color: 'var(--foreground-muted)' }}>
                {getPlanIntervalName(activePlan)}
              </Text>
            </Flex>
          )}
          {isLoading ? (
            <Skeleton />
          ) : (
            <Flex gap="small">
              <Text size={2} weight={500}>
                New cycle:
              </Text>
              <Text size={2} style={{ color: 'var(--foreground-muted)' }}>
                {getPlanIntervalName(nextPlan)} (effective from the next billing
                cycle)
              </Text>
            </Flex>
          )}
        </Flex>
        <Separator />
        <Flex justify={'end'} gap="medium" style={{ padding: 'var(--pd-16)' }}>
          <Button variant={'secondary'} onClick={cancel} size={'medium'}>
            Cancel
          </Button>
          <Button
            variant={'primary'}
            size={'medium'}
            disabled={isLoading || isCycleSwitching}
            onClick={onConfirm}
          >
            {isCycleSwitching ? 'Switching...' : 'Switch cycle'}
          </Button>
        </Flex>
      </Dialog.Content>
    </Dialog>
  );
}
