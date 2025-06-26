import { EmptyState, DataTable, Flex } from "@raystack/apsara/v1";
import type {
  V1Beta1Preference,
  V1Beta1PreferenceTrait,
} from "@raystack/frontier";

import PageHeader from "~/components/page-header";
import { getColumns } from "./columns";
import { useOutletContext } from "react-router-dom";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import styles from "./preferences.module.css";

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

  const columns = getColumns({
    traits,
    preferences,
  });

  return (
    <DataTable
      data={traits}
      columns={columns}
      mode="client"
      defaultSort={{ name: "title", order: "asc" }}
      isLoading={isPreferencesLoading}
    >
      <Flex direction="column" className={styles.tableWrapper}>
        <PageHeader
          title={pageHeader.title}
          breadcrumb={pageHeader.breadcrumb}
          className={styles.header}
        />
        <DataTable.Content
          emptyState={noDataChildren}
          classNames={{ root: styles.tableRoot }}
        />
      </Flex>
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
