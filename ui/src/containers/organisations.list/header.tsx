import { PlusIcon } from "@radix-ui/react-icons";

import {
  Button,
  DataTable,
  Flex,
  Select,
  Text,
  useTable,
} from "@raystack/apsara";
import { useNavigate } from "react-router-dom";

export type OrgStates = "enabled" | "disabled";
interface OrganizationsHeaderProps {
  stateFilterValue: OrgStates;
  onStateFilterChange: (value: OrgStates) => void;
}

export const OrganizationsHeader = ({
  stateFilterValue,
  onStateFilterChange,
}: OrganizationsHeaderProps) => {
  const navigate = useNavigate();
  const { filteredColumns, table } = useTable();
  const isFiltered = filteredColumns.length > 0;

  return (
    <>
      <Flex align="center" justify="between" style={{ padding: "16px 24px" }}>
        <Text style={{ fontSize: "14px", fontWeight: "500" }}>
          Organisations
        </Text>
        <Flex gap="small">
          <Select value={stateFilterValue} onValueChange={onStateFilterChange}>
            <Select.Trigger style={{ minWidth: "120px" }}>
              <Select.Value placeholder="Select state" />
            </Select.Trigger>
            <Select.Content>
              <Select.Item value="enabled">Enabled</Select.Item>
              <Select.Item value="disabled">Disabled</Select.Item>
            </Select.Content>
          </Select>
          {isFiltered ? <DataTable.ClearFilter /> : <DataTable.FilterOptions />}
          <DataTable.ViewOptions />
          <DataTable.GloabalSearch placeholder="Search organisations..." />
          <Button
            variant="secondary"
            onClick={() => navigate("/organisations/create")}
            style={{ width: "100%" }}
          >
            <Flex
              direction="column"
              align="center"
              style={{ paddingRight: "var(--pd-4)" }}
            >
              <PlusIcon />
            </Flex>
            new organisation
          </Button>
        </Flex>
      </Flex>
    </>
  );
};
