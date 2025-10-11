import { Flex, EmptyState } from "@raystack/apsara";
import { Outlet } from "react-router-dom";
import { useQuery } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  FrontierServiceQueries,
} from "@raystack/proton/frontier";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";

export default function PreferencesLayout() {
  const {
    data: preferencesData,
    isLoading: isPreferencesLoading,
    error: preferencesError,
    isError: isPreferencesError,
  } = useQuery(AdminServiceQueries.listPreferences, {}, {
    staleTime: 60 * 1000, // Cache for 1 minute
    refetchOnWindowFocus: false,
  });

  const {
    data: traitsData,
    isLoading: isTraitsLoading,
    error: traitsError,
    isError: isTraitsError,
  } = useQuery(FrontierServiceQueries.describePreferences, {}, {
    staleTime: 60 * 1000, // Cache for 1 minute
    refetchOnWindowFocus: false,
  });

  const preferences = preferencesData?.preferences || [];
  const traits = traitsData?.traits || [];
  const isLoading = isPreferencesLoading || isTraitsLoading;
  const isError = isPreferencesError || isTraitsError;
  const error = preferencesError || traitsError;

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
          isPreferencesLoading: isLoading,
        }}
      />
    </Flex>
  );
}
