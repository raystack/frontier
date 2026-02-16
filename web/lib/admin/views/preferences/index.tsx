import { EmptyState, DataTable, Flex } from "@raystack/apsara";
import { Preference, PreferenceTrait } from "@raystack/proton/frontier";
import { PageHeader } from "../../components/PageHeader";
import { getColumns } from "./columns";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import styles from "./preferences.module.css";

const pageHeader = {
  title: "Preferences",
  breadcrumb: [] as { name: string; href?: string }[],
};

export type PreferencesListProps = {
  preferences: Preference[];
  traits: PreferenceTrait[];
  isLoading: boolean;
};

export default function PreferencesList({
  preferences,
  traits,
  isLoading,
}: PreferencesListProps) {
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
      isLoading={isLoading}
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
