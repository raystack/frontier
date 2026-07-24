import { Button, type DataTableColumnDef } from "@raystack/apsara";
import type { ServiceUser, User } from "@raystack/proton/frontier";
import { TerminologyEntity } from "../../hooks/useTerminology";
import { useOrganizationLookup } from "../../hooks/useOrganizationLookup";
import styles from "./admins.module.css";

// The cell only has the org id — fetch the org to show its title and slug link.
const OrgCell = ({
  orgId,
  onNavigateToOrg,
}: {
  orgId: string;
  onNavigateToOrg?: (slug: string, orgId: string) => void;
}) => {
  const { data: org } = useOrganizationLookup(orgId);

  return (
    <Button
      variant="text"
      disabled={!org}
      onClick={() => org && onNavigateToOrg?.(org.name || orgId, orgId)}
      data-test-id="frontier-admin-navigate-to-org-btn"
    >
      {org?.title || org?.name || orgId}
    </Button>
  );
};

export const getColumns: (options?: {
  onNavigateToOrg?: (slug: string, orgId: string) => void;
  t?: {
    organization: TerminologyEntity;
  };
}) => DataTableColumnDef<
  User | ServiceUser,
  unknown
>[] = ({ onNavigateToOrg, t } = {}) => {
  return [
    {
      header: "Title",
      accessorKey: "title",
      classNames: {
        cell: styles["first-column"],
        header: styles["first-column"],
      },
      filterType: "string",
      enableColumnFilter: true,
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "Email",
      accessorKey: "email",
      filterType: "string",
      enableColumnFilter: true,
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "Status",
      accessorKey: "state",
      meta: {
        data: [
          { label: "Enabled", value: "enabled" },
          { label: "Disabled", value: "disabled" },
        ],
      },
      cell: (info) => info.getValue(),
      footer: (props) => props.column.id,
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
      },
    },
    {
      header: t?.organization({ case: "capital" }) || "Organization",
      accessorKey: "orgId",
      cell: (info) => {
        const org_id = info.getValue() as string;
        return org_id ? (
          <OrgCell orgId={org_id} onNavigateToOrg={onNavigateToOrg} />
        ) : (
          "-"
        );
      },
    },
  ];
};
