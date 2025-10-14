import { Skeleton, Text } from "@raystack/apsara";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import { useQuery } from "@connectrpc/connect-query";
import { memo } from "react";
import styles from "./list.module.css";

type OrganizationCellProps = { id: string };

export const OrganizationCell = memo(({ id }: OrganizationCellProps) => {
  const { data, isLoading } = useQuery(
    AdminServiceQueries.listAllOrganizations,
    {},
    {
      staleTime: 0,
      refetchOnWindowFocus: false,
    },
  );
  const orgData = data?.organizations?.find(org => org.id === id);

  if (isLoading) return <Skeleton />;
  return (
    <Text size="regular" className={styles.capitalize}>
      {orgData?.title || orgData?.name}
    </Text>
  );
});

OrganizationCell.displayName = "OrganizationCell";
