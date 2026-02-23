import { UsersList } from "./list/list";
import { UserDetailsByUserId } from "./details/user-details";

export type UsersViewProps = {
  selectedUserId?: string;
  onCloseDetail?: () => void;
  onExportUsers?: () => Promise<void>;
  onNavigateToUser?: (userId: string) => void;
  currentPath?: string;
  onNavigate?: (path: string) => void;
};

export default function UsersView({
  selectedUserId,
  onCloseDetail,
  onExportUsers,
  onNavigateToUser,
  currentPath,
  onNavigate,
}: UsersViewProps = {}) {
  if (selectedUserId) {
    return <UserDetailsByUserId userId={selectedUserId} currentPath={currentPath} onNavigate={onNavigate} />;
  }

  return (
    <UsersList
      onExportUsers={onExportUsers}
      onNavigateToUser={onNavigateToUser}
    />
  );
}
