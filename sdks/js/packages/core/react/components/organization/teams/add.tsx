import {
  InputField,
  TextField
} from '@raystack/apsara';
import { Button, toast, Image, Text, Flex, Dialog } from '@raystack/apsara/v1';

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
      <Dialog.Content style={{ padding: 0, maxWidth: '600px', width: '100%', zIndex: '60' }} overlayClassName={styles.overlay}>
        <Dialog.Header>
          <Flex justify="between">
            <Text size="large" weight="medium">
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
        </Dialog.Header>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Body>
            <Flex
              direction="column"
              gap={5}
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

                <Text size="mini" variant="danger">
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

                <Text size="mini" variant="danger">
                  {errors.name && String(errors.name?.message)}
                </Text>
              </InputField>
            </Flex>
          </Dialog.Body>
          <Dialog.Footer>
            <Button
              type="submit"
              data-test-id="frontier-sdk-add-team-btn"
              loading={isSubmitting}
              loaderText="Adding..."
            >
              Add team
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
