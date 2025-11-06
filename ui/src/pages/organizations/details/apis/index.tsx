import { DataTable, EmptyState, Flex, toast } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import styles from "./apis.module.css";
import { InfoCircledIcon } from "@radix-ui/react-icons";
import { useCallback, useContext, useEffect, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import PageTitle from "~/components/page-title";
import { getColumns } from "./columns";
import { useRQL } from "~/hooks/useRQL";
import { ServiceUserDetailsDialog } from "./details-dialog";
import { useTransport } from "@connectrpc/connect-query";
import {
  AdminService,
  SearchOrganizationServiceUsersRequestSchema,
  type SearchOrganizationServiceUsersResponse_OrganizationServiceUser,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { transformDataTableQueryToRQLRequest } from "~/utils/transform-query";

const NoCredentials = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No API Credentials found"
      subHeading="We couldn't find any matches for that keyword or filter. Try alternative terms or check for typos."
      // TODO: update icon with raystack icon
      icon={<InfoCircledIcon />}
    />
  );
};

const DEFAULT_SORT: DataTableSort = { name: "created_at", order: "desc" };

export function OrganizationApisPage() {
  const { organization, search } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";
  const transport = useTransport();
  const client = createClient(AdminService, transport);
  const [selectedServiceUser, setSelectedServiceUser] =
    useState<SearchOrganizationServiceUsersResponse_OrganizationServiceUser | null>(
      null,
    );

  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const title = `API | ${organization?.title} | Organizations`;

  const apiCallback = useCallback(
    async (apiQuery: DataTableQuery = {}) => {
      try {
        const rqlQuery = transformDataTableQueryToRQLRequest({
          ...apiQuery,
          search: searchQuery || "",
        });

        const response = await client.searchOrganizationServiceUsers(
          create(SearchOrganizationServiceUsersRequestSchema, {
            id: organizationId,
            query: rqlQuery,
          }),
        );

        return {
          organization_service_users: response.organizationServiceUsers || [],
          count: Number(response.pagination?.totalCount || 0),
        };
      } catch (error) {
        toast.error("Something went wrong", {
          description: "Unable to fetch service users",
        });
        console.error("Unable to fetch service users:", error);
        throw error;
      }
    },
    [organizationId, searchQuery, client],
  );

  const { data, loading, query, onTableQueryChange, groupCountMap, fetchMore } =
    useRQL<SearchOrganizationServiceUsersResponse_OrganizationServiceUser>({
      initialQuery: { offset: 0 },
      key: organizationId,
      dataKey: "organization_service_users",
      fn: apiCallback,
      searchParam: searchQuery || "",
      onError: (error: Error | unknown) =>
        console.error("Failed to fetch service users:", error),
    });

  useEffect(() => {
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  const isLoading = loading;

  function onDialogClose() {
    setSelectedServiceUser(null);
  }

  function onRowClick(
    row: SearchOrganizationServiceUsersResponse_OrganizationServiceUser,
  ) {
    setSelectedServiceUser(row);
  }

  const columns = getColumns({ groupCountMap });
  return (
    <Flex justify="center" className={styles["container"]}>
      <ServiceUserDetailsDialog
        serviceUser={selectedServiceUser}
        onClose={onDialogClose}
      />
      <PageTitle title={title} />
      <DataTable
        columns={columns}
        data={data}
        isLoading={isLoading}
        defaultSort={DEFAULT_SORT}
        mode="server"
        onTableQueryChange={onTableQueryChange}
        onLoadMore={fetchMore}
        query={{ ...query, search: searchQuery }}
        onRowClick={onRowClick}
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <DataTable.Toolbar />
          <DataTable.Content
            emptyState={<NoCredentials />}
            classNames={{
              table: styles["table"],
              root: styles["table-wrapper"],
              header: styles["table-header"],
            }}
          />
        </Flex>
      </DataTable>
    </Flex>
  );
}
