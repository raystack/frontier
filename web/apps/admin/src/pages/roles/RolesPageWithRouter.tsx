import { RolesPage } from "@raystack/frontier/admin";
import { useParams, useNavigate } from "react-router-dom";

export function RolesPageWithRouter() {
  const { roleId } = useParams();
  const navigate = useNavigate();

  return (
    <RolesPage
      selectedRoleId={roleId}
      onSelectRole={(id) => navigate(`/roles/${encodeURIComponent(id)}`)}
      onCloseDetail={() => navigate("/roles")}
    />
  );
}
