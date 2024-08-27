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
  const [hasMoreData, setHasMoreData] = useState(true);

  const isUserEmpty = R.either(R.isEmpty, R.isNil)(user);

  const fetchOrganizations = useCallback(async () => {
    if (!hasMoreData) return;

    setIsOrgListLoading(true);
    try {
      const [orgResp, disabledOrgResp] = await Promise.all([
        client?.adminServiceListAllOrganizations({ page_num: page, page_size: 10 }),
        client?.adminServiceListAllOrganizations({ state: "disabled", page_num: page, page_size: 10 }),
      ]);

      if (orgResp?.data?.organizations?.length) {
        setEnabledOrganizations((prev) => [
          ...prev,
          ...orgResp.data.organizations,
        ]);
      } else {
        setHasMoreData(false);
      }

      if (disabledOrgResp?.data?.organizations?.length) {
        setDisabledOrganizations((prev) => [
          ...prev,
          ...disabledOrgResp.data.organizations,
        ]);
      }
      setIsAdmin(true);
    } catch (error) {
      console.error(error);
      setIsAdmin(false);
      setHasMoreData(false);
    } finally {
      setIsOrgListLoading(false);
    }
  }, [client, page, hasMoreData]);

  const loadMoreOrganizations = () => {
    if (!isOrgListLoading && hasMoreData) {
      setPage((prevPage) => prevPage + 1);
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
