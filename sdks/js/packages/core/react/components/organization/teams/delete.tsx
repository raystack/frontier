import {
  Button,
  Checkbox,
  Dialog,
  Flex,
  Image,
  InputField,
  Separator,
  Text,
  TextField
} from '@raystack/apsara';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group } from '~/src';
import styles from '../organization.module.css';
import Skeleton from 'react-loading-skeleton';

const teamSchema = yup
  .object({
    name: yup.string()
  })
  .required();

export const DeleteTeam = () => {
  const {
    watch,
    control,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(teamSchema)
  });
  let { teamId } = useParams({ from: `/teams/$teamId/delete` });
  const navigate = useNavigate();
  const [team, setTeam] = useState<V1Beta1Group>();
  const [isTeamLoading, setIsTeamLoading] = useState(false);
  const [isAcknowledged, setIsAcknowledged] = useState(false);

  const { client, activeOrganization: organization } = useFrontier();

  useEffect(() => {
    async function getTeamDetails() {
      if (!organization?.id || !teamId) return;

      try {
        setIsTeamLoading(true);
        const {
          // @ts-ignore
          data: { group }
        } = await client?.frontierServiceGetGroup(organization?.id, teamId);
        setTeam(group);
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      } finally {
        setIsTeamLoading(false);
      }
    }
    getTeamDetails();
  }, [client, organization?.id, teamId]);

  async function onSubmit(data: any) {
    if (!organization?.id) return;
    if (!teamId) return;
    if (!client) return;

    if (data.name !== team?.name)
      return setError('name', { message: 'team name is not same' });

    try {
      await client.frontierServiceDeleteGroup(organization.id, teamId);
      toast.success('team deleted');

      navigate({ to: '/teams' });
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  const name = watch('name', '');
  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%', zIndex: '60' }}
        overlayClassname={styles.overlay}
      >
        <Flex justify="between" style={{ padding: '16px 24px' }}>
          <Text size={6} style={{ fontWeight: '500' }}>
            Verify team deletion
          </Text>
          <Image
            alt="cross"
            // @ts-ignore
            src={cross}
            onClick={() =>
              navigate({
                to: `/teams/$teamId`,
                params: {
                  teamId
                }
              })
            }
            style={{ cursor: 'pointer' }}
          />
        </Flex>
        <Separator />
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex
            direction="column"
            gap="medium"
            style={{ padding: '24px 32px' }}
          >
            {isTeamLoading ? (
              <>
                <Skeleton height={'16px'} />
                <Skeleton width={'50%'} height={'16px'} />
                <Skeleton height={'32px'} />
                <Skeleton height={'16px'} />
                <Skeleton height={'32px'} />
              </>
            ) : (
              <>
                <Text size={2}>
                  This action can not be undone. This will permanently delete
                  team <b>{team?.title}</b>.
                </Text>

                <InputField label="Please type name of the team to confirm.">
                  <Controller
                    render={({ field }) => (
                      <TextField
                        {...field}
                        // @ts-ignore
                        size="medium"
                        placeholder="Provide team name"
                      />
                    )}
                    control={control}
                    name="name"
                  />

                  <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                    {errors.name && String(errors.name?.message)}
                  </Text>
                </InputField>
                <Flex>
                  <Checkbox
                    //@ts-ignore
                    checked={isAcknowledged}
                    onCheckedChange={setIsAcknowledged}
                  ></Checkbox>
                  <Text size={2}>
                    I acknowledge I understand that all of the team data will be
                    deleted and want to proceed.
                  </Text>
                </Flex>
                <Button
                  variant="danger"
                  size="medium"
                  disabled={!name || !isAcknowledged}
                  type="submit"
                  style={{ width: '100%' }}
                  data-test-id="frontier-sdk-delete-team-btn"
                >
                  {isSubmitting ? 'Deleting...' : 'Delete this team'}
                </Button>
              </>
            )}
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
