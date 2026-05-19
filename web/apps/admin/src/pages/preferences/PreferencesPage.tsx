import { useParams, useNavigate } from "react-router-dom";
import { PreferencesView } from "@raystack/frontier/admin";
import PreferencesIcon from "~/assets/icons/preferences.svg?react";

export function PreferencesPage() {
  const { name } = useParams();
  const navigate = useNavigate();

  return (
    <PreferencesView
      selectedPreferenceName={name}
      onCloseDetail={() => navigate("/preferences")}
      onSelectPreference={(prefName: string) => navigate(`/preferences/${prefName}`)}
      icon={<PreferencesIcon />}
    />
  );
}
