'use client';

import { Image, Menu, Spinner } from '@raystack/apsara';
import deleteIcon from '../../../assets/delete.svg';

export interface DeleteProjectMenuItemProps {
  isChecking: boolean;
  onClick: () => void;
  'data-test-id': string;
}

export function DeleteProjectMenuItem({
  isChecking,
  ...props
}: DeleteProjectMenuItemProps) {
  return (
    <Menu.Item
      leadingIcon={
        isChecking ? (
          <Spinner size={2} />
        ) : (
          <Image src={deleteIcon as unknown as string} alt="Delete" width={16} height={16} />
        )
      }
      closeOnClick={false}
      disabled={isChecking}
      style={{ color: 'var(--rs-color-foreground-danger-primary)' }}
      {...props}
    >
      Delete project
    </Menu.Item>
  );
}
