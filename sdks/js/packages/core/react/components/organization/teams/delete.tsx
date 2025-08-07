import {
  Button,
  Checkbox,
  Skeleton,
  Image,
  Text,
  Flex,
  Dialog,
  toast,
  InputField
} from '@raystack/apsara';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group } from '~/src';
import styles from '../organization.module.css';

const teamSchema = yup
  .object({
    name: yup.string()
  })
  .required();

export const DeleteTeam = () => {
  const {
    watch,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
    register
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
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%' }}
        overlayClassName={styles.overlay}
      >
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="large" weight="medium">
              Verify team deletion
            </Text>
            <Image
              alt="cross"
              src={cross as unknown as string}
              onClick={() =>
                navigate({
                  to: `/teams/$teamId`,
                  params: {
                    teamId
                  }
                })
              }
              style={{ cursor: 'pointer' }}
              data-test-id="frontier-sdk-delete-team-close-btn"
            />
          </Flex>
        </Dialog.Header>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Body>
            <Flex direction="column" gap={5}>
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

                  <InputField
                    label="Please type name of the team to confirm."
                    size="large"
                    error={errors.name && String(errors.name?.message)}
                    {...register('name')}
                    placeholder="Provide team name"
                  />

                  <Flex gap="small">
                    <Checkbox
                      checked={isAcknowledged}
                      onCheckedChange={v => setIsAcknowledged(v === true)}
                      data-test-id="frontier-sdk-delete-team-checkbox"
                    />
                    <Text size={2}>
                      I acknowledge that all of the team data will be deleted
                      and want to proceed.
                    </Text>
                  </Flex>
                  <Button
                    variant="solid"
                    color="danger"
                    disabled={!name || !isAcknowledged}
                    type="submit"
                    style={{ width: '100%' }}
                    data-test-id="frontier-sdk-delete-team-btn-general"
                    loading={isSubmitting}
                    loaderText="Deleting..."
                  >
                    Delete this team
                  </Button>
                </>
              )}
            </Flex>
          </Dialog.Body>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
