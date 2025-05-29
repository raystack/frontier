import {
  Dialog,
  Flex,
  InputField,
  Text,
  TextField
} from '@raystack/apsara';
import { Button, Separator, toast, Image } from '@raystack/apsara/v1';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate } from '@tanstack/react-router';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import styles from '../organization.module.css';

const teamSchema = yup
  .object({
    title: yup.string().required(),
    name: yup
      .string()
      .required('name is a required field')
      .min(3, 'name is not valid, Min 3 characters allowed')
      .max(50, 'name is not valid, Max 50 characters allowed')
      .matches(
        /^[a-zA-Z0-9_-]{3,50}$/,
        "Only numbers, letters, '-', and '_' are allowed. Spaces are not allowed."
      )
  })
  .required();

type FormData = yup.InferType<typeof teamSchema>;

export const AddTeam = () => {
  const {
    reset,
    control,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(teamSchema)
  });
  const navigate = useNavigate({ from: '/members/modal' });
  const { client, activeOrganization: organization } = useFrontier();

  async function onSubmit(data: FormData) {
    if (!client) return;
    if (!organization?.id) return;

    try {
      await client.frontierServiceCreateGroup(organization?.id, data);
      toast.success('Team added');
      navigate({ to: '/teams' });
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
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
            Add Team
          </Text>
          <Image
            alt="cross"
            src={cross as unknown as string}
            onClick={() => navigate({ to: '/teams' })}
            style={{ cursor: 'pointer' }}
            data-test-id="frontier-sdk-add-team-close-btn"
          />
        </Flex>
        <Separator />
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex
            direction="column"
            gap="medium"
            style={{ padding: '24px 32px' }}
          >
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

              <Text size={1} style={{ color: 'var(--rs-color-foreground-danger-primary)' }}>
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
                    placeholder="Provide team name"
                  />
                )}
                control={control}
                name="name"
              />

              <Text size={1} style={{ color: 'var(--rs-color-foreground-danger-primary)' }}>
                {errors.name && String(errors.name?.message)}
              </Text>
            </InputField>
          </Flex>
          <Separator />
          <Flex align="end" style={{ padding: 'var(--rs-space-5)' }}>
            <Button
              type="submit"
              data-test-id="frontier-sdk-add-team-btn"
              loading={isSubmitting}
              loaderText="Adding..."
            >
              Add team
            </Button>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
