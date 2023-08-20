import {
  Button,
  Flex,
  InputField,
  Separator,
  Text,
  TextField
} from '@raystack/apsara';

import { yupResolver } from '@hookform/resolvers/yup';
import { useCallback, useEffect } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group, V1Beta1Organization } from '~/src';

const teamSchema = yup
  .object({
    title: yup.string().required(),
    name: yup.string().required()
  })
  .required();

interface GeneralTeamProps {
  team?: V1Beta1Group;
  organization?: V1Beta1Organization;
}

export const General = ({ organization, team }: GeneralTeamProps) => {
  const {
    reset,
    control,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(teamSchema)
  });

  let { teamId } = useParams();
  const { client } = useFrontier();

  useEffect(() => {
    reset(team);
  }, [reset, team]);

  async function onSubmit(data: any) {
    if (!client) return;
    if (!organization?.id) return;
    if (!teamId) return;

    try {
      await client.frontierServiceUpdateGroup(organization?.id, teamId, data);
      toast.success('Team updated');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return (
    <Flex direction="column" gap="large" style={{ paddingTop: '32px' }}>
      <form onSubmit={handleSubmit(onSubmit)}>
        <Flex direction="column" gap="medium" style={{ maxWidth: '320px' }}>
          <InputField label="Team title">
            <Controller
              render={({ field }) => (
                <TextField
                  {...field}
                  // @ts-ignore
                  size="medium"
                  placeholder="Provide team title"
                />
              )}
              control={control}
              name="title"
            />

            <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
              {errors.title && String(errors.title?.message)}
            </Text>
          </InputField>
          <InputField label="Team name">
            <Controller
              render={({ field }) => (
                <TextField
                  {...field}
                  // @ts-ignore
                  size="medium"
                  disabled
                  placeholder="Provide team name"
                />
              )}
              control={control}
              name="name"
            />

            <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
              {errors.title && String(errors.title?.message)}
            </Text>
          </InputField>
          <Button variant="primary" size="medium" type="submit">
            {isSubmitting ? 'updating...' : 'Update team'}
          </Button>
        </Flex>
      </form>
      <Separator />
      <GeneralDeleteTeam organization={organization} />
      <Separator />
    </Flex>
  );
};

export const GeneralDeleteTeam = ({ organization }: GeneralTeamProps) => {
  const { client } = useFrontier();
  let { teamId } = useParams();
  const navigate = useNavigate();
  const {
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm();

  const organizationId = organization?.id;

  const onDeleteOrganization = useCallback(async () => {
    if (!organizationId || !teamId) return;
    try {
      await client?.frontierServiceDeleteGroup(organizationId, teamId);
      navigate('/teams');
    } catch ({ error }: any) {
      console.log(error);
      toast.error('Something went wrong', {
        description: `${error.message}`
      });
    }
  }, [client, navigate, organizationId, teamId]);

  return (
    <Flex direction="column" gap="medium">
      <Text size={3} style={{ color: 'var(--foreground-muted)' }}>
        If you want to permanently delete this team and all of its data.
      </Text>
      <Button
        variant="danger"
        type="submit"
        size="medium"
        onClick={() => navigate(`/teams/${teamId}/delete`)}
      >
        {isSubmitting ? 'deleting...' : 'Delete team'}
      </Button>
    </Flex>
  );
};
