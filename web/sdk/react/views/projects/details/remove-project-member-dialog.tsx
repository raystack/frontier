'use client';

import { Button, Flex, Text, toast, Image, Dialog } from '@raystack/apsara';
import {
    useMutation,
    createConnectQueryKey,
    useTransport
} from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import {
    FrontierServiceQueries,
    RemoveProjectMemberRequestSchema,
    ListProjectUsersRequestSchema,
    ListProjectGroupsRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import cross from '~/react/assets/cross.svg';
import orgStyles from '../../../components/organization/organization.module.css';

export interface RemoveProjectMemberDialogProps {
    open: boolean;
    onOpenChange?: (value: boolean) => void;
    projectId: string;
    memberId: string;
    memberType: 'user' | 'group';
}

export const RemoveProjectMemberDialog = ({
    open,
    onOpenChange,
    projectId,
    memberId,
    memberType
}: RemoveProjectMemberDialogProps) => {
    const queryClient = useQueryClient();
    const transport = useTransport();

    const handleOpenChange = (value: boolean) => {
        onOpenChange?.(value);
    };

    const { mutateAsync: removeProjectMember, isPending } = useMutation(
        FrontierServiceQueries.removeProjectMember,
        {
            onError: (err: Error) =>
                toast.error('Something went wrong', { description: err.message })
        }
    );

    async function onConfirm() {
        try {
            await removeProjectMember(
                create(RemoveProjectMemberRequestSchema, {
                    projectId: projectId,
                    principalId: memberId,
                    principalType: memberType === 'group' ? 'app/group' : 'app/user'
                })
            );
            if (projectId) {
                await queryClient.invalidateQueries({
                    queryKey: createConnectQueryKey({
                        schema: FrontierServiceQueries.listProjectUsers,
                        transport,
                        input: create(ListProjectUsersRequestSchema, {
                            id: projectId,
                            withRoles: true
                        }),
                        cardinality: 'finite'
                    })
                });
                await queryClient.invalidateQueries({
                    queryKey: createConnectQueryKey({
                        schema: FrontierServiceQueries.listProjectGroups,
                        transport,
                        input: create(ListProjectGroupsRequestSchema, {
                            id: projectId,
                            withRoles: true
                        }),
                        cardinality: 'finite'
                    })
                });
            }
            handleOpenChange(false);
            toast.success('Member removed');
        } catch (error) {
            console.error('Failed to remove member:', error);
        }
    }

    return (
        <Dialog open={open} onOpenChange={handleOpenChange}>
            <Dialog.Content
                style={{ padding: 0, maxWidth: '400px', width: '100%' }}
                overlayClassName={orgStyles.overlay}
            >
                <Dialog.Header>
                    <Flex justify="between" align="center" style={{ width: '100%' }}>
                        <Text size="large" weight="medium">
                            Remove project member
                        </Text>
                        <Image
                            data-test-id="frontier-sdk-remove-project-member-close-btn"
                            alt="cross"
                            src={cross as unknown as string}
                            onClick={() => handleOpenChange(false)}
                            style={{ cursor: 'pointer' }}
                        />
                    </Flex>
                </Dialog.Header>

                <Dialog.Body>
                    <Flex direction="column" gap={5}>
                        <Text size="regular">
                            Are you sure you want to remove this member from the project?
                        </Text>
                    </Flex>
                </Dialog.Body>

                <Dialog.Footer>
                    <Flex justify="end" gap={5}>
                        <Button
                            size="normal"
                            color="neutral"
                            variant="outline"
                            onClick={() => handleOpenChange(false)}
                            data-test-id="frontier-sdk-remove-project-member-cancel-btn"
                            disabled={isPending}
                        >
                            Cancel
                        </Button>
                        <Button
                            size="normal"
                            color="danger"
                            variant="solid"
                            onClick={onConfirm}
                            data-test-id="frontier-sdk-remove-project-member-confirm-btn"
                            disabled={isPending}
                            loading={isPending}
                            loaderText="Removing"
                        >
                            Remove
                        </Button>
                    </Flex>
                </Dialog.Footer>
            </Dialog.Content>
        </Dialog>
    );
};
