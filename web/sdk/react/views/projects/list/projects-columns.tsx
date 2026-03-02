'use client';

import {
    DotsHorizontalIcon,
    Pencil1Icon,
    TrashIcon
} from '@radix-ui/react-icons';
import { Text, DropdownMenu } from '@raystack/apsara';
import type { Project } from '@raystack/proton/frontier';
import type { DataTableColumnDef } from '@raystack/apsara';
import orgStyles from '../../../components/organization/organization.module.css';

export const getColumns = (
    userAccessOnProject: Record<string, string[]>,
    onProjectClick?: (projectId: string) => void,
    onDeleteProjectClick?: (projectId: string) => void
): DataTableColumnDef<Project, unknown>[] => [
    {
        header: 'Title',
        accessorKey: 'title',
        cell: ({ row, getValue }) => {
            return (
                <span
                    onClick={() => onProjectClick?.(row.original.id || '')}
                    style={{
                        textDecoration: 'none',
                        color: 'var(--rs-color-foreground-base-primary)',
                        fontSize: 'var(--rs-font-size-small)',
                        cursor: 'pointer'
                    }}
                >
                    {getValue() as string}
                </span>
            );
        }
    },
    {
        header: 'Members',
        accessorKey: 'membersCount',
        cell: ({ row, getValue }) => {
            const value = getValue() as string;
            return value ? <Text>{value} members</Text> : null;
        }
    },
    {
        header: '',
        accessorKey: 'id',
        meta: {
            style: {
                textAlign: 'end'
            }
        },
        enableSorting: false,
        cell: ({ row, getValue }) => (
            <ProjectActions
                project={row.original as Project}
                userAccessOnProject={userAccessOnProject}
                onProjectClick={onProjectClick}
                onDeleteProjectClick={onDeleteProjectClick}
            />
        )
    }
];

const ProjectActions = ({
    project,
    userAccessOnProject,
    onProjectClick,
    onDeleteProjectClick
}: {
    project: Project;
    userAccessOnProject: Record<string, string[]>;
    onProjectClick?: (projectId: string) => void;
    onDeleteProjectClick?: (projectId: string) => void;
}) => {
    const canUpdateProject = (userAccessOnProject[project.id!] ?? []).includes(
        'update'
    );
    const canDeleteProject = (userAccessOnProject[project.id!] ?? []).includes(
        'delete'
    );
    const canDoActions = canUpdateProject || canDeleteProject;

    return canDoActions ? (
        <DropdownMenu placement="bottom-end">
            <DropdownMenu.Trigger asChild style={{ cursor: 'pointer' }}>
                <DotsHorizontalIcon />
            </DropdownMenu.Trigger>
            {/* @ts-ignore */}
            <DropdownMenu.Content portal={false}>
                <DropdownMenu.Group>
                    {canUpdateProject ? (
                        <DropdownMenu.Item
                            onClick={() => onProjectClick?.(project.id || '')}
                            className={orgStyles.dropdownActionItem}
                        >
                            <Pencil1Icon /> Rename
                        </DropdownMenu.Item>
                    ) : null}
                    {canDeleteProject ? (
                        <DropdownMenu.Item
                            onClick={() => onDeleteProjectClick?.(project.id || '')}
                            className={orgStyles.dropdownActionItem}
                        >
                            <TrashIcon /> Delete project
                        </DropdownMenu.Item>
                    ) : null}
                </DropdownMenu.Group>
            </DropdownMenu.Content>
        </DropdownMenu>
    ) : null;
};

