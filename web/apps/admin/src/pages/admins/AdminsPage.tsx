import { AdminsView, useAdminPaths } from "@raystack/frontier/admin";
import { useNavigate } from "react-router-dom";

export function AdminsPage() {
  const navigate = useNavigate();
  const paths = useAdminPaths();

  return (
    <AdminsView
      onNavigateToOrg={(orgId: string) => navigate(`/${paths.organizations}/${orgId}`)}
    />
  );
}
