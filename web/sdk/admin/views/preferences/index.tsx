import { EmptyState, DataTable, Flex } from "@raystack/apsara";
import type { ReactNode } from "react";
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
  onSelectPreference?: (name: string) => void;
  icon?: ReactNode;
};

export default function PreferencesList({
  preferences,
  traits,
  isLoading,
  onSelectPreference,
  icon,
}: PreferencesListProps) {
  const columns = getColumns({
    traits,
    preferences,
    onSelectPreference,
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
          icon={icon}
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
