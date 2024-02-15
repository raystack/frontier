import {
  DataTable,
  EmptyState,
  Flex,
  Switch,
  Text,
  TextField,
} from "@raystack/apsara";
import { V1Beta1Preference, V1Beta1PreferenceTrait } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import PageHeader from "~/components/page-header";
import { getColumns } from "./columns";

const pageHeader = {
  title: "Preferences",
  breadcrumb: [],
};

export default function PreferencesList() {
  const { client } = useFrontier();
  const [preferences, setPreferences] = useState<V1Beta1Preference[]>([]);
  const [traits, setTraits] = useState<V1Beta1PreferenceTrait[]>([]);
  const [isPreferencesLoading, setIsPreferencesLoading] = useState(false);

  useEffect(() => {
    async function getPreferences() {
      try {
        setIsPreferencesLoading(true);
        const [traitResp, valuesMapResp] = await Promise.all([
          client?.frontierServiceDescribePreferences(),
          client?.adminServiceListPreferences(),
        ]);

        if (valuesMapResp?.data?.preferences) {
          setPreferences(valuesMapResp?.data?.preferences);
        }
        if (traitResp?.data?.traits) {
          setTraits(traitResp?.data?.traits);
        }
      } catch (err) {
        console.error(err);
      } finally {
        setIsPreferencesLoading(false);
      }
    }
    getPreferences();
  }, []);

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
    isLoading: isPreferencesLoading,
  });

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        // @ts-ignore
        data={data}
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <PageHeader
            title={pageHeader.title}
            breadcrumb={pageHeader.breadcrumb}
          />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
}

export const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 traits</h3>
    <div className="pera">Try creating new traits.</div>
  </EmptyState>
);
