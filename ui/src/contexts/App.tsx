import * as R from "ramda";
import React, {
  PropsWithChildren,
  createContext,
  useCallback,
  useEffect,
  useState,
} from "react";
import {
  V1Beta1ListPlatformUsersResponse,
  V1Beta1Organization,
  V1Beta1Plan,
} from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";

// Setting this to 1000 initially till APIs support filters and sorting.
const page_size = 1000 

type OrgMap = Record<string, V1Beta1Organization>;

interface AppContextValue {
  orgMap: OrgMap;
  isAdmin: boolean;
  isLoading: boolean;
  organizations: V1Beta1Organization[];
  plans: V1Beta1Plan[];
  platformUsers?: V1Beta1ListPlatformUsersResponse;
  fetchPlatformUsers: () => void;
  loadMoreOrganizations: () => void;
}

const AppContextDefaultValue = {
  orgMap: {},
  isAdmin: false,
  isLoading: false,
  organizations: [],
  plans: [],
  platformUsers: {
    users: [],
    serviceusers: [],
  },
  fetchPlatformUsers: () => {},
  loadMoreOrganizations: () => {}
};

export const AppContext = createContext<AppContextValue>(
  AppContextDefaultValue
);

export const AppContextProvider: React.FC<PropsWithChildren> = function ({
  children,
}) {
  const { client, user, isUserLoading } = useFrontier();
  const [isOrgListLoading, setIsOrgListLoading] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);
  const [enabledOrganizations, setEnabledOrganizations] = useState<
    V1Beta1Organization[]
  >([]);
  const [disabledOrganizations, setDisabledOrganizations] = useState<
    V1Beta1Organization[]
  >([]);
  const [plans, setPlans] = useState<V1Beta1Plan[]>([]);
  const [isPlansLoading, setIsPlansLoading] = useState(false);

  const [isPlatformUsersLoading, setIsPlatformUsersLoading] = useState(false);
  const [platformUsers, setPlatformUsers] =
    useState<V1Beta1ListPlatformUsersResponse>();

  const [page, setPage] = useState(1);
  const [enabledOrgHasMoreData, setEnabledOrgHasMoreData] = useState(true);
  const [disabledOrgHasMoreData, setDisabledOrgHasMoreData] = useState(true);

  const isUserEmpty = R.either(R.isEmpty, R.isNil)(user);

  const fetchOrganizations = useCallback(async () => {
    if (!enabledOrgHasMoreData && !disabledOrgHasMoreData) return;

    setIsOrgListLoading(true);
    try {
      const [orgResp, disabledOrgResp] = await Promise.all([
        client?.adminServiceListAllOrganizations({ page_num: page, page_size }),
        client?.adminServiceListAllOrganizations({ state: "disabled", page_num: page, page_size }),
      ]);

      if (orgResp?.data?.organizations?.length) {
        setEnabledOrganizations((prev: V1Beta1Organization[]) => [
          ...prev,
          ...(orgResp.data.organizations || []),
        ]);
      } else {
        setEnabledOrgHasMoreData(false);
      }

      if (disabledOrgResp?.data?.organizations?.length) {
        setDisabledOrganizations((prev: V1Beta1Organization[]) => [
          ...prev,
          ...(disabledOrgResp.data.organizations || []),
        ]);
      } else {
        setDisabledOrgHasMoreData(false);
      }
      setIsAdmin(true);
    } catch (error) {
      console.error(error);
      setIsAdmin(false);
      setEnabledOrgHasMoreData(false);
      setDisabledOrgHasMoreData(false);
    } finally {
      setIsOrgListLoading(false);
    }
  }, [client, page, enabledOrgHasMoreData, disabledOrgHasMoreData]);

  const loadMoreOrganizations = () => {
    if (!isOrgListLoading && (enabledOrgHasMoreData || disabledOrgHasMoreData)) {
      setPage((prevPage: number) => prevPage + 1);
    }
  };

  useEffect(() => {
    if (!isUserEmpty) {
      fetchOrganizations();
    }
  }, [client, isUserEmpty, page, fetchOrganizations]);

  const fetchPlatformUsers = useCallback(async () => {
    setIsPlatformUsersLoading(true);
    try {
      const resp = await client?.adminServiceListPlatformUsers();
      if (resp?.data) {
        setPlatformUsers(resp?.data);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setIsPlatformUsersLoading(false);
    }
  }, [client]);

  useEffect(() => {
    async function getPlans() {
      setIsPlansLoading(true);
      try {
        const resp = await client?.frontierServiceListPlans();
        const planList = resp?.data?.plans || [];
        setPlans(planList);
      } catch (error) {
        console.error(error);
      } finally {
        setIsPlansLoading(false);
      }
    }

    if (isAdmin) {
      getPlans();
      fetchPlatformUsers();
    }
  }, [client, isAdmin, fetchPlatformUsers]);

  const isLoading =
    isOrgListLoading ||
    isUserLoading ||
    isPlansLoading ||
    isPlatformUsersLoading;
  const organizations = [...enabledOrganizations, ...disabledOrganizations];

  const orgMap = organizations.reduce((acc, org) => {
    const orgId = org?.id || "";
    if (orgId) acc[orgId] = org;
    return acc;
  }, {} as OrgMap);

  return (
    <AppContext.Provider
      value={{
        orgMap,
        isLoading,
        isAdmin,
        organizations,
        plans,
        platformUsers,
        fetchPlatformUsers,
        loadMoreOrganizations,
      }}
    >
      {children}
    </AppContext.Provider>
  );
};
