import { Flex, EmptyState } from "@raystack/apsara";
import { Outlet } from "react-router-dom";
import { createQueryOptions, useTransport } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  FrontierServiceQueries,
  Preference,
  PreferenceTrait,
  ListPreferencesResponse,
  DescribePreferencesResponse,
} from "@raystack/proton/frontier";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { useQueries } from "@tanstack/react-query";
import type { ConnectError } from "@connectrpc/connect";

export default function PreferencesLayout() {
  const transport = useTransport();

  const [preferencesQuery, traitsQuery] = useQueries({
    queries: [
      {
        ...createQueryOptions(AdminServiceQueries.listPreferences, {}, { transport }),
        staleTime: Infinity,
      },
      {
        ...createQueryOptions(FrontierServiceQueries.describePreferences, {}, { transport }),
        staleTime: Infinity,
      },
    ],
  });

  const preferences = ((preferencesQuery.data as ListPreferencesResponse)?.preferences || []) as Preference[];
  const traits = ((traitsQuery.data as DescribePreferencesResponse)?.traits || []) as PreferenceTrait[];
  const isLoading = preferencesQuery.isLoading || traitsQuery.isLoading;
  const isError = preferencesQuery.isError || traitsQuery.isError;
  const error = (preferencesQuery.error || traitsQuery.error) as ConnectError | null;

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <Flex direction="row" style={{ height: "100%", width: "100%" }}>
        <EmptyState
          icon={<ExclamationTriangleIcon />}
          heading="Error Loading Preferences"
          subHeading={
            error?.message ||
            "Something went wrong while loading preferences. Please try again."
          }
        />
      </Flex>
    );
  }

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <Outlet
        context={{
          preferences,
          traits,
          isLoading,
        }}
      />
    </Flex>
  );
}
