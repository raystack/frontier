import { UsersView, useAdminPaths } from "@raystack/frontier/admin";
import { useCallback } from "react";
import { useParams, useNavigate, useLocation } from "react-router-dom";
import { clients } from "~/connect/clients";
import { exportCsvFromStream } from "~/utils/helper";

const adminClient = clients.admin({ useBinary: true });

export function UsersPage() {
  const { userId } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const paths = useAdminPaths();

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
