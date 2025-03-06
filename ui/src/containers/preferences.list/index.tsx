import { DataTable } from "@raystack/apsara";
import { EmptyState } from "@raystack/apsara/v1";
import { V1Beta1Preference, V1Beta1PreferenceTrait } from "@raystack/frontier";

import PageHeader from "~/components/page-header";
import { getColumns } from "./columns";
import { useOutletContext } from "react-router-dom";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";

const pageHeader = {
  title: "Preferences",
  breadcrumb: [],
};

interface ContextType {
  preferences: V1Beta1Preference[];
  traits: V1Beta1PreferenceTrait[];
  isPreferencesLoading: boolean;
}

export function usePreferences() {
  return useOutletContext<ContextType>();
}

export default function PreferencesList() {
  const { preferences, traits, isPreferencesLoading } = usePreferences();

  const tableStyle = traits?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const data = isPreferencesLoading
    ? [...new Array(5)].map((_, i) => ({
        name: i,
        title: "",
      }))
    : traits;

  const columns = getColumns({
    traits,
    preferences,
  });

  return (
    <DataTable
      // @ts-ignore
      data={data}
      columns={columns}
      emptyState={noDataChildren}
      parentStyle={{ height: "calc(100vh - 60px)" }}
      style={tableStyle}
      isLoading={isPreferencesLoading}
    >
      <DataTable.Toolbar>
        <PageHeader
          title={pageHeader.title}
          breadcrumb={pageHeader.breadcrumb}
        />
        <DataTable.FilterChips style={{ padding: "8px 24px" }} />
      </DataTable.Toolbar>
    </DataTable>
  );
}

export const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="No traits"
    subHeading="Try creating new traits."
  />
);
