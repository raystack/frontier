import { UsersList } from "./list/list";
import { UserDetailsByUserId } from "./details/user-details";

export type UsersViewProps = {
  selectedUserId?: string;
  onCloseDetail?: () => void;
  onExportUsers?: () => Promise<void>;
  onNavigateToUser?: (userId: string) => void;
};

export default function UsersView({
  selectedUserId,
  onCloseDetail,
  onExportUsers,
  onNavigateToUser,
}: UsersViewProps = {}) {
  if (selectedUserId) {
    return <UserDetailsByUserId userId={selectedUserId} />;
  }

  return (
    <UsersList
      onExportUsers={onExportUsers}
      onNavigateToUser={onNavigateToUser}
    />
  );
}
