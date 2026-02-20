import { UsersView } from "@raystack/frontier/admin";
import { useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { clients } from "~/connect/clients";
import { exportCsvFromStream } from "~/utils/helper";

const adminClient = clients.admin({ useBinary: true });

export function UsersPage() {
  const { userId } = useParams();
  const navigate = useNavigate();

  const onExportUsers = useCallback(async () => {
    await exportCsvFromStream(adminClient.exportUsers, {}, "users.csv");
  }, []);

  const onNavigateToUser = useCallback(
    (id: string) => {
      navigate(`/users/${id}/security`);
    },
    [navigate],
  );

  return (
    <UsersView
      selectedUserId={userId}
      onCloseDetail={() => navigate("/users")}
      onExportUsers={onExportUsers}
      onNavigateToUser={onNavigateToUser}
    />
  );
}
