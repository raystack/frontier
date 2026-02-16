import { RolesView } from "@raystack/frontier/admin";
import { useParams, useNavigate } from "react-router-dom";

export function RolesPage() {
  const { roleId } = useParams();
  const navigate = useNavigate();

  return (
    <RolesView
      selectedRoleId={roleId}
      onSelectRole={(id) => navigate(`/roles/${encodeURIComponent(id)}`)}
      onCloseDetail={() => navigate("/roles")}
    />
  );
}
