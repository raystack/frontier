import { useParams, useNavigate } from "react-router-dom";
import { PreferencesView } from "@raystack/frontier/admin";

export function PreferencesPage() {
  const { name } = useParams();
  const navigate = useNavigate();

  return (
    <PreferencesView
      selectedPreferenceName={name}
      onCloseDetail={() => navigate("/preferences")}
    />
  );
}
