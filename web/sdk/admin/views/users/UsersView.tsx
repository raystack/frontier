import { UsersList } from "./list/list";
import { UserDetailsByUserId } from "./details/user-details";

export type UsersViewProps = {
  /** When set, renders the user detail view for this user ID instead of the list. */
  selectedUserId?: string;
  /** Called when the detail view is closed. Use to navigate back to the users list. */
  onCloseDetail?: () => void;
  /** Callback to export users list as CSV. Shown in navbar when provided. */
  onExportUsers?: () => Promise<void>;
  /** Called when a user row is clicked. Use to navigate to the user detail page. */
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
