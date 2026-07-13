import { UsersView, useAdminPaths } from "@raystack/frontier/admin";
import { useCallback, useEffect } from "react";
import { useParams, useNavigate, useLocation } from "react-router-dom";
import { clients } from "~/connect/clients";
import { exportCsvFromStream } from "~/utils/helper";

const adminClient = clients.admin({ useBinary: true });

export default function UsersPage() {
  const { userId } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const paths = useAdminPaths();

  // Security is the default tab for the user detail page: redirect the bare
  // `/users/:userId` index to `/users/:userId/security` so the tab stays selected.
  useEffect(() => {
    const securityPath = `/${paths.users}/${userId}/security`;
    if (userId && location.pathname !== securityPath) {
      navigate(securityPath, { replace: true });
    }
  }, [userId, location.pathname, navigate, paths.users]);

  const onExportUsers = useCallback(async () => {
    await exportCsvFromStream(adminClient.exportUsers, {}, "users.csv");
  }, []);

  const onNavigateToUser = useCallback(
    (id: string) => {
      navigate(`/${paths.users}/${id}/security`);
    },
    [navigate, paths.users],
  );

  return (
    <UsersView
      selectedUserId={userId}
      onCloseDetail={() => navigate(`/${paths.users}`)}
      onExportUsers={onExportUsers}
      onNavigateToUser={onNavigateToUser}
      currentPath={location.pathname}
      onNavigate={navigate}
    />
  );
}
