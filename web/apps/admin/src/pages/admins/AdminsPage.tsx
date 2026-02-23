import { AdminsView } from "@raystack/frontier/admin";
import { useNavigate } from "react-router-dom";

export function AdminsPage() {
  const navigate = useNavigate();

  return (
    <AdminsView
      onNavigateToOrg={(orgId: string) => navigate(`/organizations/${orgId}`)}
    />
  );
}
