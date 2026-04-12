'use client';

import {
    Button,
    toast,
    Image,
    Text,
    Flex,
    Dialog,
    InputField
} from '@raystack/apsara';

import { yupResolver } from '@hookform/resolvers/yup';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useMutation } from '@connectrpc/connect-query';
import {
    FrontierServiceQueries,
    CreateGroupRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import orgStyles from '../../../components/organization/organization.module.css';
import { handleConnectError } from '~/utils/error';

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

export interface AddTeamDialogProps {
    open: boolean;
    onOpenChange: (value: boolean) => void;
}

export const AddTeamDialog = ({ open, onOpenChange }: AddTeamDialogProps) => {
    const {
        handleSubmit,
        setError,
        formState: { errors, isSubmitting },
        register,
        reset
    } = useForm({
        resolver: yupResolver(teamSchema)
    });
    const { activeOrganization: organization } = useFrontier();

    const { mutateAsync: createTeam } = useMutation(
        FrontierServiceQueries.createGroup,
        {
            onSuccess: () => {
                toast.success('Team added');
                reset();
                onOpenChange(false);
            }
        }
    );

    async function onSubmit(data: FormData) {
        if (!organization?.id) return;

        const request = create(CreateGroupRequestSchema, {
            orgId: organization.id,
            body: {
                title: data.title,
                name: data.name
            }
        });

        try {
            await createTeam(request);
        } catch (error) {
            handleConnectError(error, {
                AlreadyExists: () => setError('name', {
                    message: 'A team with this name already exists. Please use a different name.'
                }),
                PermissionDenied: () => toast.error('You don\'t have permission to perform this action'),
                InvalidArgument: (err) => toast.error('Invalid input', { description: err.message }),
                Default: (err) => toast.error('Something went wrong', { description: err.message }),
            });
        }
    }

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <Dialog.Content
                style={{ padding: 0, maxWidth: '600px', width: '100%' }}
                overlayClassName={orgStyles.overlay}
            >
                <Dialog.Header>
                    <Flex justify="between" align="center" style={{ width: '100%' }}>
                        <Text size="large" weight="medium">
                            Add Team
                        </Text>
                        <Image
                            alt="cross"
                            src={cross as unknown as string}
                            onClick={() => onOpenChange(false)}
                            style={{ cursor: 'pointer' }}
                            data-test-id="frontier-sdk-add-team-close-btn"
                        />
                    </Flex>
                </Dialog.Header>
                <form onSubmit={handleSubmit(onSubmit)}>
                    <Dialog.Body>
                        <Flex direction="column" gap={5}>
                            <InputField
                                label="Team title"
                                size="large"
                                error={errors.title && String(errors.title?.message)}
                                {...register('title')}
                                placeholder="Provide team title"
                            />
                            <InputField
                                label="Team name"
                                size="large"
                                error={errors.name && String(errors.name?.message)}
                                {...register('name')}
                                placeholder="Provide team name"
                            />
                        </Flex>
                    </Dialog.Body>
                    <Dialog.Footer>
                        <Flex align="end">
                            <Button
                                type="submit"
                                data-test-id="frontier-sdk-add-team-btn"
                                loading={isSubmitting}
                                loaderText="Adding..."
                            >
                                Add team
                            </Button>
                        </Flex>
                    </Dialog.Footer>
                </form>
            </Dialog.Content>
        </Dialog>
    );
};

